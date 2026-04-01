# Developer Guide

Welcome to the **go-service-template** developer guide. This document covers all aspects of local development, from initial setup to deploying changes.

## Table of Contents

1. [Initial Setup](#initial-setup)
2. [Development Workflow](#development-workflow)
3. [Code Structure](#code-structure)
4. [Testing](#testing)
5. [Debugging](#debugging)
6. [Common Tasks](#common-tasks)

---

## Initial Setup

### Prerequisites

Before starting, install:
- **Go 1.25+** — [Download](https://go.dev/dl/)
- **Docker Desktop** — [Download](https://www.docker.com/products/docker-desktop)
- **Git** — [Download](https://git-scm.com/)
- **VS Code** (optional but recommended) — [Download](https://code.visualstudio.com/)

### One-Time Setup

```bash
# 1. Clone and enter the repository
git clone https://github.com/your-org/your-service my-service
cd my-service

# 2. Copy environment configuration
cp .env.example .env

# 3. If using VS Code: install dependencies and recommended extensions
make vscode-setup

# 4. Start all local services (database, cache, monitoring)
make docker-run
```

This runs the service at **http://localhost:8080** with full observability stack:
- **API**: http://localhost:8080
- **Health**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics (Prometheus format)
- **Prometheus UI**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin / admin)

### Local Dev (Without Docker)

If you prefer running the service without Docker:

```bash
# Terminal 1: Start PostgreSQL and Redis
# Option A: Docker containers only
docker run --rm -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:16
docker run --rm -d -p 6379:6379 redis:7

# Option B: If installed locally
postgres &
redis-server &

# Terminal 2: Run migrations then the service
make migrate-up
make run
```

---

## Development Workflow

### Running the Service

```bash
# Quick development server with hot reload (requires air)
make run

# Build optimized binary
make build
./bin/server

# Run with debug logging
APP_LOG_LEVEL=debug make run
```

### Making Code Changes

The repository follows a **layered architecture**:

```
Handler (HTTP) → Service (Business Logic) → Repository (Data Access)
```

When adding a feature:

1. **Define the domain model** in `internal/service/example.go`
2. **Write the repository method** in `internal/repository/`
3. **Write the business logic** in `internal/service/`
4. **Write the HTTP handler** in `internal/handler/example.go`
5. **Add the route** in `internal/handler/routes.go`
6. **Write tests** in `*_test.go` files
7. **Update OpenAPI spec** in `api/openapi.yaml`
8. **Run migrations** if schema changed

### Code Formatting & Linting

```bash
# Format all code
make fmt

# Run linter (catches style issues)
make lint

# Run type checker
make vet

# Run all code quality checks
make fmt lint vet

# Check for security vulnerabilities
make vuln
```

### Testing

```bash
# Run all tests with race detector
make test

# Run with coverage
go test ./... -race -cover

# Run specific package
go test ./internal/service -v

# Run with verbose output
go test -v ./...
```

See [TESTING.md](TESTING.md) for detailed testing strategies.

### Database Migrations

```bash
# Create a new migration
migrate create -ext sql -dir migrations -seq create_users_table

# Run pending migrations (up)
make migrate-up

# Rollback last migration (down)
make migrate-down

# See ./scripts/migrate.sh for more options
```

---

## Code Structure

### Directory Layout

```
internal/
├── handler/          # HTTP layers (parsing, response)
│   ├── routes.go     # Route definitions
│   ├── example.go    # Endpoint implementations
│   └── health.go     # Health checks
├── service/          # Business logic & caching
│   ├── service.go    # Interface + constructor
│   └── example.go    # Domain types & methods
└── repository/       # Data access (SQL layer)
    ├── postgres.go   # GORM implementation (writes)
    └── pgx.go        # pgx/v5 implementation (reads)

config/              # Configuration loading from env
api/                 # OpenAPI specification
deploy/              # Docker, Kubernetes, Prometheus
migrations/          # SQL migration files
scripts/             # Utility scripts
```

### Naming Conventions

| Item | Convention | Example |
|---|---|---|
| Package | lowercase, short | `handler`, `service`, `repo` |
| Type | PascalCase | `User`, `CreateRequest`, `ListParams` |
| Method | camelCase | `CreateUser()`, `ListByEmail()` |
| Variable | camelCase | `userID`, `createdAt` |
| Constant | SCREAMING_SNAKE_CASE | `MAX_RETRIES`, `DEFAULT_TIMEOUT` |
| Table | snake_case, plural | `users`, `user_sessions` |
| Column | snake_case | `created_at`, `email_verified` |

### Design Patterns

#### Cache-Aside Pattern

The service layer implements cache-aside (look-aside) caching:

```go
// Repository call with optional caching
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    // Try cache first
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        return cached, nil
    }
    
    // Fetch from database
    user, err := s.repo.Get(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Populate cache for next time
    _ = s.cache.Set(ctx, cacheKey, user, ttl)
    return user, nil
}
```

**Never cache:**
- Authentication/authorization tokens
- Real-time data (weather, stock prices)
- One-time values (OTPs, reset codes)
- User-specific sensitive data

#### Dependency Injection

All layers receive dependencies via constructor:

```go
// ✓ Good: Dependencies injected
func NewHandler(svc Service, log *zap.Logger) *Handler {
    return &Handler{svc: svc, log: log}
}

// ✗ Avoid: Global singletons
var GlobalService = NewService()
```

#### Error Handling

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Use structured logging
log.Error("database error", zap.Error(err), zap.String("user_id", id))

// Consider retry logic for transient errors
```

---

## Testing

### Test Structure

```go
// File: internal/service/example_test.go
func TestGetExample(t *testing.T) {
    svc := New(fakeRepo, fakeCache, logger)
    
    result, err := svc.Get(context.Background(), "id-123")
    
    assert.NoError(t, err)
    assert.Equal(t, "id-123", result.ID)
}
```

### Test Categories

1. **Unit Tests** — Test single functions in isolation using mocks
   ```bash
   go test ./internal/service -v
   ```

2. **Integration Tests** — Test multiple components together with a real database
   ```bash
   docker compose -f deploy/docker-compose.test.yml up
   go test -tags=integration ./...
   ```

See [TESTING.md](TESTING.md) for comprehensive testing guide.

---

## Debugging

### Enable Debug Logs

```bash
# Set log level to debug
APP_LOG_LEVEL=debug make run

# Or in .env
APP_LOG_LEVEL=debug
```

### Remote Debugging (Delve)

```bash
# Install Delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Run service with debugger
dlv debug ./cmd/server

# In Delve REPL:
# (dlv) break main.main    # Set breakpoint
# (dlv) continue           # Run to breakpoint
# (dlv) next               # Step over line
# (dlv) print var          # Print variable value
# (dlv) quit               # Exit debugger
```

### VS Code Debugging

1. Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go)
2. Add `.vscode/launch.json`:
   ```json
   {
     "version": "0.2.0",
     "configurations": [
       {
         "name": "Connect to Delve",
         "type": "go",
         "mode": "remote",
         "remotePath": "",
         "port": 2345,
         "host": "127.0.0.1",
         "showLog": true
       }
     ]
   }
   ```
3. Start Delve: `dlv debug ./cmd/server --headless --listen=:2345`
4. Press F5 in VS Code to attach

---

## Common Tasks

### Add a New Endpoint

1. **Create domain type** in `internal/service/example.go`:
   ```go
   type CreateUserRequest struct {
       Name  string `json:"name"`
       Email string `json:"email"`
   }
   ```

2. **Add repository method** in `internal/repository/postgres.go`:
   ```go
   func (r *Repository) CreateUser(ctx context.Context, u *User) error {
       return r.db.WithContext(ctx).Create(u).Error
   }
   ```

3. **Add service method** in `internal/service/service.go`:
   ```go
   func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
       user := &User{Name: req.Name, Email: req.Email}
       return user, s.repo.CreateUser(ctx, user)
   }
   ```

4. **Add handler** in `internal/handler/example.go`:
   ```go
   func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
       var req service.CreateUserRequest
       if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
           http.Error(w, err.Error(), http.StatusBadRequest)
           return
       }
       user, err := h.svc.CreateUser(r.Context(), &req)
       // ... response handling
   }
   ```

5. **Add route** in `internal/handler/routes.go`:
   ```go
   mux.Post("/users", h.CreateUser)
   ```

6. **Update OpenAPI spec** in `api/openapi.yaml`

7. **Write tests** in `internal/handler/example_test.go`, etc.

### Update Configuration

1. Add new env var to `.env.example`
2. Add field to `config.Config` struct in `config/config.go`
3. Use in code — config is already loaded and available
4. Document in README.md Configuration table

### Run All Code Quality Checks

```bash
make fmt lint vet test vuln
```

### Generate API Documentation

```bash
# The OpenAPI spec is already in api/openapi.yaml
# To view it, run:
docker run -p 80:8080 \
  -e SWAGGER_JSON=/foo/openapi.yaml \
  -v $(pwd)/api:/foo \
  swaggerapi/swagger-ui
```

Then visit **http://localhost** to view API documentation.

---

## VS Code Integration

See [VS_CODE_SETUP.md](VS_CODE_SETUP.md) for detailed guidance on:
- Setting up the Go extension
- Configuring MCP servers for AI assistance
- Pre-commit hooks and code quality automation

---

## Questions?

- Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues
- Review test examples in `*_test.go` files
- Read [ARCHITECTURE.md](ARCHITECTURE.md) for design patterns
- See [CONTRIBUTING.md](CONTRIBUTING.md) for PR guidelines
