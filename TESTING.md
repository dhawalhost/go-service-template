# Testing Guide

This document outlines testing strategies, patterns, and best practices for the go-service-template.

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test Structure](#test-structure)
3. [Unit Testing](#unit-testing)
4. [Integration Testing](#integration-testing)
5. [Test Patterns](#test-patterns)
6. [Coverage Goals](#coverage-goals)
7. [Running Tests](#running-tests)

---

## Testing Philosophy

This repository follows the **test pyramid**:

```
       /\
      /  \  E2E Tests (UI, API integration) — 5-10%
     /____\

      /\
     /  \  Integration Tests (DB, Redis) — 20-30%
    /____\

      /\
     /  \  Unit Tests (services, handlers) — 60-70%
    /____\
```

**Principle:** Test business logic thoroughly with fast unit tests. Use integration tests for critical paths. E2E tests verify the whole system works.

### Testing Goals

- ✓ Catch bugs before production
- ✓ Enable safe refactoring
- ✓ Document expected behavior
- ✓ Verify edge cases and error conditions
- ✓ Keep test suite fast (< 5 seconds for unit tests)

---

## Test Structure

### File Organization

```
internal/
├── handler/
│   ├── handler.go
│   ├── example.go
│   ├── example_test.go      ← Tests for example.go
│   ├── handler_test.go      ← Tests for handler.go
│   └── routes.go
├── service/
│   ├── service.go
│   ├── service_test.go
│   ├── example.go
│   └── example_test.go
└── repository/
    ├── repository.go
    ├── postgres.go
    ├── postgres_test.go     ← Integration tests (DB)
    └── pgx.go
```

### Test File Template

```go
package service_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "go.uber.org/zap"
    
    "github.com/dhawalhost/go-service-template/internal/service"
)

func TestGetUser(t *testing.T) {
    // Arrange
    ctx := context.Background()
    mockRepo := &mockRepository{}
    mockCache := &mockCache{}
    svc := service.New(mockRepo, mockCache, zap.NewNop())
    
    // Act
    result, err := svc.GetUser(ctx, "test-id")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "test-id", result.ID)
}
```

---

## Unit Testing

Unit tests are **fast**, **independent**, and **deterministic**. They test a single function or method in isolation using **mocks** for dependencies.

### Mocking Pattern

Create a `*_test.go` file with mock implementations:

```go
// service_test.go

type mockRepository struct {
    GetUserFunc func(ctx context.Context, id string) (*User, error)
    CreateUserFunc func(ctx context.Context, u *User) error
}

func (m *mockRepository) GetUser(ctx context.Context, id string) (*User, error) {
    if m.GetUserFunc != nil {
        return m.GetUserFunc(ctx, id)
    }
    return nil, errors.New("not mocked")
}

func (m *mockRepository) CreateUser(ctx context.Context, u *User) error {
    if m.CreateUserFunc != nil {
        return m.CreateUserFunc(ctx, u)
    }
    return errors.New("not mocked")
}
```

**Or use testify's `mock` package:**

```go
import "github.com/stretchr/testify/mock"

type mockRepository struct {
    mock.Mock
}

func (m *mockRepository) GetUser(ctx context.Context, id string) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

// In test:
func TestGetUser(t *testing.T) {
    mockRepo := new(mockRepository)
    mockRepo.On("GetUser", mock.Anything, "123").
        Return(&User{ID: "123", Name: "Alice"}, nil)
    
    svc := service.New(mockRepo, &mockCache{}, zap.NewNop())
    user, err := svc.GetUser(context.Background(), "123")
    
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
    mockRepo.AssertCalled(t, "GetUser", mock.Anything, "123")
}
```

### Testing Handlers

```go
package handler_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "go.uber.org/zap"
    
    "github.com/dhawalhost/go-service-template/internal/handler"
    "github.com/dhawalhost/go-service-template/internal/service"
)

type mockService struct {
    CreateUserFunc func(ctx context.Context, req *service.CreateUserRequest) (*service.User, error)
}

func (m *mockService) CreateUser(ctx context.Context, req *service.CreateUserRequest) (*service.User, error) {
    if m.CreateUserFunc != nil {
        return m.CreateUserFunc(ctx, req)
    }
    return nil, errors.New("not mocked")
}

func TestCreateUserHandler(t *testing.T) {
    // Arrange
    mockSvc := &mockService{
        CreateUserFunc: func(ctx context.Context, req *service.CreateUserRequest) (*service.User, error) {
            return &service.User{
                ID:    "123",
                Name:  req.Name,
                Email: req.Email,
            }, nil
        },
    }
    
    h := handler.New(mockSvc, zap.NewNop())
    
    body := service.CreateUserRequest{
        Name:  "Alice",
        Email: "alice@example.com",
    }
    bodyBytes, _ := json.Marshal(body)
    
    req := httptest.NewRequest("POST", "/users", bytes.NewReader(bodyBytes))
    w := httptest.NewRecorder()
    
    // Act
    h.CreateUser(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var responseBody service.User
    json.NewDecoder(w.Body).Decode(&responseBody)
    assert.Equal(t, "Alice", responseBody.Name)
}
```

### Table-Driven Tests

For testing multiple scenarios:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name      string
        email     string
        wantErr   bool
        wantMsg   string
    }{
        {"valid email", "user@example.com", false, ""},
        {"missing @", "user", true, "must contain @"},
        {"empty", "", true, "must not be empty"},
        {"spaces", "user @example.com", true, "must not contain spaces"},
        {"no domain", "user@", true, "must have domain"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
            }
            
            if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.wantMsg) {
                t.Errorf("ValidateEmail(%q) error message = %q, want %q", tt.email, err.Error(), tt.wantMsg)
            }
        })
    }
}
```

### Testing Error Cases

Always test both success and failure paths:

```go
func TestCreateUser_ValidationError(t *testing.T) {
    mockRepo := &mockRepository{}
    svc := service.New(mockRepo, &mockCache{}, zap.NewNop())
    
    _, err := svc.CreateUser(context.Background(), &CreateUserRequest{
        Name:  "",  // Invalid: empty name
        Email: "alice@example.com",
    })
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "name must not be empty")
}

func TestCreateUser_DatabaseError(t *testing.T) {
    mockRepo := &mockRepository{
        CreateUserFunc: func(ctx context.Context, u *User) error {
            return errors.New("database connection failed")
        },
    }
    svc := service.New(mockRepo, &mockCache{}, zap.NewNop())
    
    _, err := svc.CreateUser(context.Background(), &CreateUserRequest{
        Name:  "Alice",
        Email: "alice@example.com",
    })
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to create user")
}
```

---

## Integration Testing

Integration tests run against real services (database, cache) to verify components work together.

### Build Tag Pattern

Mark integration tests with build tags:

```go
//go:build integration
// +build integration

package repository_test

import (
    "context"
    "testing"
    
    "github.com/dhawalhost/go-service-template/internal/repository"
)

func TestCreateUser_Integration(t *testing.T) {
    t.Skip("Run with: go test -tags=integration ./...")
    
    // Test against real database
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repository.NewPostgres(db)
    user := &User{Name: "Alice", Email: "alice@example.com"}
    
    err := repo.CreateUser(context.Background(), user)
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)  // ID auto-generated
}
```

### Running Integration Tests

```bash
# Run all tests (excluding integration tests by default)
make test

# Run only integration tests
go test -tags=integration ./...

# Run specific integration test
go test -tags=integration -run TestCreateUser_Integration ./internal/repository
```

### Test Database Setup

```go
func setupTestDB(t *testing.T) *gorm.DB {
    dsn := "postgres://user:password@localhost:5432/test_db"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to connect to test database: %v", err)
    }
    
    // Run migrations
    err = db.AutoMigrate(&User{}, &Order{})
    if err != nil {
        t.Fatalf("failed to migrate: %v", err)
    }
    
    // Clean data before test
    db.Exec("TRUNCATE users, orders RESTART IDENTITY")
    
    return db
}
```

### Docker Compose for Integration Tests

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: test_user
      POSTGRES_PASSWORD: test_password
      POSTGRES_DB: test_db
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "test_user"]
      interval: 1s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "PING"]
      interval: 1s
      timeout: 3s
      retries: 5
```

Run before tests:
```bash
docker-compose -f docker-compose.test.yml up
go test -tags=integration ./...
docker-compose -f docker-compose.test.yml down
```

---

## Test Patterns

### Arrange-Act-Assert

Every test follows this structure:

```go
func TestExample(t *testing.T) {
    // ARRANGE: Set up test data and dependencies
    mockRepo := &mockRepository{...}
    svc := service.New(mockRepo, &mockCache{}, zap.NewNop())
    
    // ACT: Call the code under test
    result, err := svc.DoSomething(ctx, "input")
    
    // ASSERT: Verify the result
    assert.NoError(t, err)
    assert.Equal(t, expectedValue, result)
}
```

### Subtests

Use subtests to group related tests:

```go
func TestService_UserOperations(t *testing.T) {
    t.Run("creates user successfully", func(t *testing.T) {
        // Test create
    })
    
    t.Run("retrieves existing user", func(t *testing.T) {
        // Test get
    })
    
    t.Run("returns error for non-existent user", func(t *testing.T) {
        // Test not found
    })
}
```

### Fixtures

Reusable test data:

```go
func newTestUser(t *testing.T) *User {
    return &User{
        ID:       "test-id-123",
        Name:     "Test User",
        Email:    "test@example.com",
        CreateAt: time.Now(),
    }
}

func TestGetUser(t *testing.T) {
    user := newTestUser(t)
    mockRepo := &mockRepository{
        GetUserFunc: func(ctx context.Context, id string) (*User, error) {
            return user, nil
        },
    }
    // ...
}
```

---

## Coverage Goals

### Coverage Targets

```go
// internal/service/ (business logic)
// TARGET: 80%+ coverage
// HIGH PRIORITY: Core business rules, validation, error handling

// internal/handler/ (HTTP layer)
// TARGET: 70%+ coverage
// MEDIUM PRIORITY: Route handling, JSON parsing, response formatting

// internal/repository/ (data access)
// TARGET: Unit tests via mocks (50%+)
//         Integration tests cover real scenario
// MEDIUM PRIORITY: SQL execution tested via integration tests
```

### Measuring Coverage

```bash
# Generate coverage report
go test ./... -cover -coverprofile=coverage.out

# View in browser
go tool cover -html=coverage.out

# View coverage by function
go tool cover -func=coverage.out

# Check coverage threshold
go tool cover -func=coverage.out | awk -F'\t' '{sum+=$(NF-1); count++} END {print (sum/count) "%"}'
```

### Exclude from Coverage

Mark untestable code:

```go
//nolint:coverage  // Server is tested via integration tests
func RunServer(cfg *Config) error {
    // ...
}
```

---

## Running Tests

### Quick Checks

```bash
# Unit tests only
make test

# With verbose output
go test -v ./...

# Specific package
go test -v ./internal/service

# Specific test
go test -v -run TestCreateUser ./internal/service
```

### Watch Mode (Requires `air`)

```bash
# Auto-run tests on file changes
air --cmd "go test -v ./..." --poll

# Or with a custom binary:
go run github.com/cosmtrek/air@latest

# Test on save (with build excluded)
```

### Continuous Integration

```bash
# CI script: scripts/test.sh
#!/bin/bash
set -e
echo "Running tests..."
go test -v ./... -race -cover
echo "Checking coverage..."
go test ./... -coverprofile=coverage.out
coverage=$(go tool cover -func=coverage.out | awk '{sum+=$3; count++} END {print int(sum/count) "%"}')
echo "Coverage: $coverage"
```

### Test Parallelization

Tests run in parallel by default:

```bash
# Serialize tests (if they conflict)
go test -v -p 1 ./...

# Run with specific concurrency
go test -v -p 4 ./...

# Parallel timeout
go test -timeout 60s -p 4 ./...
```

---

## Mocking Best Practices

### When to Mock

✓ **Do mock:**
- Database repositories (integration tests separately)
- External APIs (use `httptest.Server` instead of mocking the client)
- Expensive operations (file I/O, network)
- Time-dependent code (use `time.Time` arguments instead of `time.Now()`)

✗ **Don't mock:**
- Simple types (strings, ints)
- Standard library functions
- Code you own and want to test
- The thing you're testing

### Spy Pattern

Sometimes you want to track calls without affecting behavior:

```go
type spyRepository struct {
    *realRepository
    getCallCount int
}

func (s *spyRepository) GetUser(ctx context.Context, id string) (*User, error) {
    s.getCallCount++
    return s.realRepository.GetUser(ctx, id)
}

func TestServiceCallsRepository(t *testing.T) {
    realRepo := realRepository{}
    spy := &spyRepository{realRepository: &realRepo}
    
    svc := service.New(spy, &mockCache{}, zap.NewNop())
    svc.DoSomething(context.Background())
    
    assert.Greater(t, spy.getCallCount, 0)
}
```

---

## Assertions Library

Use `github.com/stretchr/testify/assert`:

```go
assert.NoError(t, err)
assert.Error(t, err)
assert.Equal(t, expected, actual)
assert.NotEqual(t, x, y)
assert.True(t, condition)
assert.False(t, condition)
assert.Nil(t, value)
assert.NotNil(t, value)
assert.Contains(t, container, element)
assert.Len(t, slice, 5)
assert.Empty(t, slice)
assert.NotEmpty(t, slice)
```

Or use `require` for assertions that stop test immediately:

```go
require.NoError(t, err)  // Fails test and stops execution
assert.NoError(t, err)   // Logs failure but continues
```

---

## Summary

- ✓ Unit test business logic with mocks
- ✓ Table-driven tests for multiple scenarios
- ✓ Integration tests for critical paths
- ✓ 80%+ coverage on service layer, 70%+ on handler layer
- ✓ Fast, independent, deterministic tests
- ✓ Clear Arrange-Act-Assert structure
- ✓ Test both success and failure paths
- ✓ Mock external dependencies

See [CONTRIBUTING.md](CONTRIBUTING.md) for testing requirements for pull requests.
