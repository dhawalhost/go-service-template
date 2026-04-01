# Architecture Guide

This document describes the design patterns, architectural decisions, and best practices used throughout the go-service-template.

## Table of Contents

1. [Layered Architecture](#layered-architecture)
2. [Design Patterns](#design-patterns)
3. [Dependency Injection](#dependency-injection)
4. [Error Handling](#error-handling)
5. [Caching Strategy](#caching-strategy)
6. [Database Access Patterns](#database-access-patterns)
7. [Observability](#observability)

---

## Layered Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Clients / Users                       │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP Request
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  Global Middleware (RequestID, Logger, Recovery, CORS, etc.)  │
│                    (gokit.Middleware)                         │
└────────────────────────┬────────────────────────────────────┘
                         │ *http.Request with context
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Handler Layer                              │
│              (internal/handler/*)                             │
│  ├─ Parse HTTP request → domain types                         │
│  ├─ Validate input                                            │
│  ├─ Call service layer                                        │
│  └─ Write HTTP response (JSON, status codes, headers)         │
└────────────────────────┬────────────────────────────────────┘
                         │ Domain types (User, CreateRequest)
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                              │
│              (internal/service/*)                             │
│  ├─ Business logic & domain rules                             │
│  ├─ Cache-aside pattern for reads                             │
│  ├─ Transaction coordination                                  │
│  ├─ Complex workflow orchestration                            │
│  └─ Authorization checks                                      │
└────────────────────────┬────────────────────────────────────┘
                         │ Domain types (filtered/computed)
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  Repository Layer                             │
│              (internal/repository/*)                          │
│  ├─ postgres.go (GORM) — Create, Update, Delete              │
│  └─ pgx.go (pgx/v5) — Optimized List & Get queries           │
└────────────────────────┬────────────────────────────────────┘
                         │ SQL queries
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              PostgreSQL + Redis + Files                       │
└─────────────────────────────────────────────────────────────┘
```

### Why This Structure?

**Separation of Concerns:**
- **Handler** — Translation layer between HTTP and business domain
- **Service** — Pure business logic, testable without HTTP
- **Repository** — Data access, testable with mocks or integration test DB

**Testability:**
- Mock service in handler tests
- Mock repository in service tests
- Mock database in repository tests

**Reusability:**
- Services can be called from handlers OR other services OR CLI commands
- Business logic is decoupled from HTTP

---

## Design Patterns

### 1. Dependency Injection

**Pattern:** Constructor-based injection

All layers receive dependencies via constructor, never via global state:

```go
// ✓ Good: Service receives its dependencies
func (s *Service) __init__(repo Repository, cache Cache, log *zap.Logger) {
    s.repo = repo
    s.cache = cache
    s.log = log
}

// ✗ Bad: Service uses global singleton
var GlobalRepository = createRepository()

func (s *Service) DoSomething() {
    user, _ := GlobalRepository.GetUser()
}
```

**Benefits:**
- Easy to mock for testing
- Clear dependency graph
- No hidden dependencies
- Easier to reason about code

**In cmd/server/main.go:**

```go
// Wire dependency graph: repo → service → handler
repo := repository.NewPostgres(db.GORM)
svc := service.New(repo, redisCache, log)
h := handler.New(svc, log)
```

### 2. Interfaces Over Concrete Types

**Pattern:** Define interfaces in the layer that uses them

```go
// In service/service.go (service layer defines what it needs)
type Repository interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, u *User) error
    ListUsers(ctx context.Context, params ListParams) ([]*User, error)
}

// In handler/handler.go (handler defines what it needs)
type Service interface {
    CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
}

// In repository/postgres.go (concrete implementation)
type PostgresRepository struct {
    db *gorm.DB
}

func (r *PostgresRepository) GetUser(ctx context.Context, id string) (*User, error) {
    // GORM implementation
}
```

**Benefits:**
- Decouple layers
- Easy to swap implementations (e.g., cache layer during testing)
- Clear contracts

### 3. Cache-Aside Pattern (Look-Aside)

**Pattern:** Check cache before hitting database

```go
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", id)
    
    // Check cache first
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        if user, ok := cached.(*User); ok {
            return user, nil
        }
    }
    
    // Miss: fetch from database
    user, err := s.repo.GetUser(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Populate cache for next request (fire-and-forget)
    _ = s.cache.Set(ctx, cacheKey, user, 5*time.Minute)
    
    return user, nil
}
```

**When use cache-aside:**
- ✓ Frequently read data (users, settings, products)
- ✓ Data that's expensive to compute
- ✓ Data that changes infrequently

**Never cache:**
- ✗ Authentication tokens (use dedicated session store instead)
- ✗ One-time values (OTPs, password reset codes)
- ✗ Real-time data (stock prices, location tracking)
- ✗ User-specific sensitive data if other users share cache

**Cache Invalidation:**

When data changes, invalidate the cache:

```go
func (s *Service) UpdateUser(ctx context.Context, u *User) error {
    // Update database
    if err := s.repo.UpdateUser(ctx, u); err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:%s", u.ID)
    _ = s.cache.Delete(ctx, cacheKey)
    
    return nil
}
```

### 4. Read/Write Splitting

**Pattern:** Different implementations for reads vs writes

```
Handler Layer
    ↓
Service Layer (orchestration)
    ├→ repository.Get() [OPTIMIZED FOR READS — pgx/v5]
    ├→ repository.Create() [TRANSACTIONAL — GORM]
    └→ repository.List() [OPTIMIZED AGGREGATION — pgx/v5]

Repository Layer
    ├─ postgres.go (GORM)   → Handles writes, transactions
    └─ pgx.go (pgx/v5)      → Handles optimized reads
```

**Why separate?**

| Operation | Tool | Why |
|---|---|---|
| CREATE, UPDATE, DELETE | GORM | Handles transactions, associations, hooks |
| GET, LIST | pgx/v5 | Faster, lower memory, custom SQL |

See [Database Guide](#database-access-patterns) for details.

---

## Dependency Injection

### The Wiring Pattern

In `cmd/server/main.go`, we construct and wire the dependency graph once:

```go
// Load configuration
cfg := svcconfig.MustLoad()

// Init dependencies (order matters!)
log, _ := logger.New(cfg.Log.Level, cfg.Log.Development)
db, _ := database.New(ctx, cfg.Database)
redisCache, _ := cache.NewRedis(cfg.Redis)

// Set up observability
observability.InitMetrics(cfg.Telemetry.ServiceName)

// Wire: repo → service → handler
repo := repository.NewPostgres(db.GORM)
svc := service.New(repo, redisCache, log)
h := handler.New(svc, log)

// Start server
srv := server.New(cfg.Server, h.NewRouter())
srv.ListenAndServe()
```

### Testing With Dependency Injection

Because of DI, testing is straightforward:

```go
// Test: Service with mocked repository
func TestGetUser(t *testing.T) {
    mockRepo := &mockRepository{
        GetUserFunc: func(ctx context.Context, id string) (*User, error) {
            return &User{ID: id, Name: "Alice"}, nil
        },
    }
    mockCache := &mockCache{}
    
    svc := service.New(mockRepo, mockCache, zap.NewNop())
    user, err := svc.GetUser(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}
```

---

## Error Handling

### Wrapping Errors with Context

Always wrap errors to provide context:

```go
// ✓ Good: Wrapped error with context
if err != nil {
    return fmt.Errorf("failed to create user in database: %w", err)
}

// ✗ Bad: Error suppressed
if err != nil {
    return nil
}

// ✗ Bad: No context
if err != nil {
    return err
}
```

### Structured Logging

Use structured fields for error context:

```go
// ✓ Good: Structured error logging
if err := s.repo.CreateUser(ctx, user); err != nil {
    h.log.Error(
        "failed to create user",
        zap.Error(err),
        zap.String("user_id", user.ID),
        zap.String("email", user.Email),
        zap.Duration("request_duration", time.Since(start)),
    )
    return nil, fmt.Errorf("failed to create user: %w", err)
}

// ✗ Bad: Unstructured error logging
if err != nil {
    fmt.Println("Error:", err)
    return nil, err
}
```

### HTTP Error Responses

Handlers should convert domain errors to appropriate HTTP status codes:

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req service.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // 400: Invalid input
        http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
        h.log.Error("failed to decode request", zap.Error(err))
        return
    }
    
    user, err := h.svc.CreateUser(r.Context(), &req)
    if err != nil {
        // 500: Server error (should never expose internal error to client)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        h.log.Error("failed to create user", zap.Error(err))
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

---

## Caching Strategy

### Cache Keys

Use structured, namespaced keys:

```go
// Pattern: entity_type:id:variant
cacheKey := fmt.Sprintf("user:%s", userID)           // Simple entity
cacheKey := fmt.Sprintf("user:%s:full", userID)      // With details
cacheKey := fmt.Sprintf("users:list:active")         // Collection
cacheKey := fmt.Sprintf("session:%s:token", sessID)  // Related data

// Avoid:
cacheKey := userID                    // Not namespaced
cacheKey := "user" + userID           // Hard to parse
```

### Cache TTLs

Different TTLs for different data stability:

```go
// Data that rarely changes
const userTTL = 24 * time.Hour       // Cache for a day

// Data that changes moderately
const sessionTTL = 1 * time.Hour     // Cache for an hour

// Computed data
const aggregateTTL = 5 * time.Minute // Cache for 5 minutes

// Data expected to change soon
const settingsTTL = 30 * time.Second // Cache briefly
```

### Cache Warming

For critical data, pre-load cache:

```go
func (s *Service) init() error {
    // Pre-load frequently accessed data
    systems, err := s.repo.GetActiveSystems(ctx)
    if err != nil {
        s.log.Warn("failed to warm cache", zap.Error(err))
        return nil  // Non-fatal, cache will be loaded on first access
    }
    
    for _, sys := range systems {
        key := fmt.Sprintf("system:%s", sys.ID)
        _ = s.cache.Set(context.Background(), key, sys, 1*time.Hour)
    }
    
    return nil
}
```

---

## Database Access Patterns

### GORM (for writes)

Use GORM for CREATE, UPDATE, DELETE operations:

```go
// Create
func (r *PostgresRepository) CreateUser(ctx context.Context, u *User) error {
    return r.db.WithContext(ctx).Create(u).Error
}

// Update
func (r *PostgresRepository) UpdateUser(ctx context.Context, u *User) error {
    return r.db.WithContext(ctx).Save(u).Error
}

// Delete
func (r *PostgresRepository) DeleteUser(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&User{}, "id = ?", id).Error
}

// Transaction
func (r *PostgresRepository) TransferCredits(ctx context.Context, from, to string, amount int) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Model(&User{}).Where("id = ?", from).Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
            return err
        }
        return tx.Model(&User{}).Where("id = ?", to).Update("balance", gorm.Expr("balance + ?", amount)).Error
    }).Error
}
```

**Benefits of GORM:**
- Automatic transaction handling
- Built-in support for hooks (BeforeCreate, AfterUpdate, etc.)
- Association management (has_many, belongs_to, etc.)
- Migration support

### pgx/v5 (for reads)

Use pgx/v5 for optimized SELECT queries:

```go
func (r *pgxRepository) GetUser(ctx context.Context, id string) (*User, error) {
    var u User
    err := r.pool.QueryRow(ctx,
        "SELECT id, name, email, created_at FROM users WHERE id = $1",
        id,
    ).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    
    if err == pgx.ErrNoRows {
        return nil, ErrNotFound
    }
    return &u, err
}

// For aggregations
func (r *pgxRepository) ListSalesReport(ctx context.Context, startDate, endDate time.Time) ([]SalesReport, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT 
            DATE(created_at) as date,
            COUNT(*) as total_orders,
            SUM(amount) as total_revenue
        FROM orders
        WHERE created_at BETWEEN $1 AND $2
        GROUP BY DATE(created_at)
        ORDER BY date DESC
    `, startDate, endDate)
    // ... scan rows
}
```

**Benefits of pgx/v5:**
- Lower overhead (direct database driver)
- Faster for high-volume reads
- Custom SQL for complex queries
- Connection pooling optimized for throughput

### Choosing Between GORM and pgx

| Operation | Use | Reason |
|---|---|---|
| Simple GET by ID | pgx | Faster, less overhead |
| Complex filters/joins | pgx | Custom SQL is clearer |
| CREATE/UPDATE/DELETE | GORM | Transaction & hook support |
| Transaction | GORM | Built-in transaction support |
| Aggregation/Report | pgx | Custom optimized SQL |
| Bulk operations | GORM | Easier API for associations |

---

## Observability

### Structured Logging

Always use structured fields with `zap`:

```go
h.log.Info("user created successfully",
    zap.String("user_id", user.ID),
    zap.String("email", user.Email),
    zap.Time("created_at", user.CreatedAt),
)
```

**Standard fields for all logs:**
- Request ID (via middleware)
- User ID (if available)
- Operation result (success/failure)
- Timing information for slow operations

### Metrics

The service automatically exports Prometheus metrics via `gokit.observability`:

```
# Automatically available metrics:

# HTTP request metrics
http_request_duration_seconds{method, path, status}
http_requests_total{method, path, status}

# Database metrics (from pgxpool)
db_query_duration_seconds{operation}

# Redis metrics
redis_command_duration_seconds{command}
redis_commands_total{command, status}

# Custom metrics
custom_metric_name{label1, label2}
```

**Accessing metrics:**
```bash
curl http://localhost:8080/metrics
```

### Distributed Tracing

Optional OpenTelemetry tracing for request flows:

```bash
# Enable via environment:
APP_TELEMETRY_ENABLED=true
APP_TELEMETRY_OTLP_ENDPOINT=http://localhost:4318
```

Traces show the full request flow: handler → service → repository → database.

---

## Summary

**The architecture embraces:**
- ✓ Layered separation for testability
- ✓ Dependency injection for flexibility
- ✓ Clear interfaces between layers
- ✓ Cache-aside pattern for performance
- ✓ Read/write splitting for optimization
- ✓ Structured logging for observability
- ✓ Error wrapping for context

This keeps the codebase maintainable, testable, and scalable.
