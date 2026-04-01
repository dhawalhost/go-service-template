# Troubleshooting Guide

Common issues and solutions for developing with go-service-template.

## Table of Contents

1. [Setup Issues](#setup-issues)
2. [Development Issues](#development-issues)
3. [Testing Issues](#testing-issues)
4. [Deployment Issues](#deployment-issues)
5. [Performance Issues](#performance-issues)
6. [Getting Help](#getting-help)

---

## Setup Issues

### `make vscode-setup` fails on macOS

**Error:** `command not found: uv` or `command not found: node`

**Solution:**

```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Try setup again
make vscode-setup

# If still fails, install manually
brew install uv node

# Then run setup
make vscode-setup
```

### VS Code extensions not appearing

**Error:** Extensions list is empty in VS Code

**Solution:**

1. Verify setup completed: `make vscode-setup-help`
2. Reload VS Code window: `Cmd+R`
3. Manually search and install:
   - `golang.go`
   - `github.copilot`
   - `github.copilot-chat`
   - `ms-azuretools.vscode-docker`
   - `redhat.vscode-yaml`

### Copilot not authenticating

**Error:** "No authorization. Sign in to GitHub..."

**Solution:**

1. Click Copilot icon in status bar (bottom right)
2. Click "Sign in to use GitHub Copilot"
3. Browser opens → approve access
4. Return to VS Code → confirmation code appears
5. If code doesn't appear: `Cmd+Shift+P` → "GitHub Copilot: Sign In"

### `code` CLI not found

**Error:** `command not found: code` when running setup script

**Solution:**

1. Open VS Code
2. `Cmd+Shift+P` (or `F1`)
3. Search "Shell Command: Install 'code' command in PATH"
4. Click to install
5. Restart terminal
6. Run `make vscode-setup` again

### Docker not running

**Error:** `Cannot connect to Docker daemon`

**Solution:**

```bash
# Start Docker Desktop (macOS)
open -a Docker

# Or start Docker daemon (Linux)
sudo systemctl start docker

# Verify Docker is running
docker ps

# Then try again
make docker-run
```

---

## Development Issues

### Go version mismatch

**Error:** `go: go.mod file indicates Go 1.25, but we are running Go X.Y`

**Solution:**

```bash
# Check your Go version
go version

# If too old, upgrade Go
# Visit https://go.dev/dl/ and install

# Or use brew
brew upgrade go
```

### Import errors after adding dependency

**Error:** `cannot find module` or `no Go files in directory`

**Solution:**

```bash
# Download and tidy modules
go mod tidy

# Verify module download
go mod download

# Verify imports are correct
go vet ./...
```

### Linter fails with timeout

**Error:** `Timeout: context deadline exceeded`

**Solution:**

```bash
# Increase timeout in Makefile
make lint  # Already has 15m timeout

# Or manually with longer timeout
golangci-lint run --timeout=20m ./...

# Clean linter cache
golangci-lint cache clean
```

### Port already in use

**Error:** `listen tcp :8080: bind: address already in use`

**Solution:**

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process (get PID from above)
kill -9 <PID>

# Or use different port
APP_SERVER_ADDR=:8081 make run

# Or stop Docker containers
docker compose -f deploy/docker-compose.yml down
```

### Redis connection refused

**Error:** `REFUSED — connect: connection refused`

**Solution:**

```bash
# Start Docker stack (includes Redis)
make docker-run

# Or manually start Redis
redis-server

# Or verify it's running
redis-cli ping
# Should return: PONG
```

### Database connection errors

**Error:** `could not connect to server`, `failed to connect to postgres`

**Solution:**

```bash
# Verify DATABASE_DSN in .env
cat .env | grep DATABASE_DSN

# Test connection manually
psql "$APP_DATABASE_DSN" -c "SELECT 1"

# Start Docker stack
make docker-run

# Check if Postgres container is healthy
docker ps | grep postgres

# View PostgreSQL logs
docker compose -f deploy/docker-compose.yml logs postgres
```

### Migrations not running

**Error:** `ERROR: relation "users" does not exist`

**Solution:**

```bash
# Run migrations manually
make migrate-up

# Verify migrations ran
psql "$APP_DATABASE_DSN" -c "\dt"  # List tables

# If still failing, check migration files exist
ls -la migrations/

# Check migration status
# (Install migrate-cli if needed)
migrate -path migrations -database "$APP_DATABASE_DSN" version
```

### Hot reload not working

**Error:** Code changes don't reload; have to restart server manually

**Solution:**

The template doesn't include hot reload by default. Use these options:

```bash
# Option 1: Use `air` for hot reload
go install github.com/cosmtrek/air@latest
air

# Option 2: Use `watchmedo` (requires watchdog)
pip install watchdog
watchmedo auto-restart -d . -p '*.go' -- go run ./cmd/server

# Option 3: Manual restart
# Edit code, then Ctrl+C and `make run` again
```

---

## Testing Issues

### `go test` timeout

**Error:** `context deadline exceeded` when running tests

**Solution:**

```bash
# Increase timeout
go test -timeout 60s ./...

# Run tests in parallel with longer timeout
go test -timeout 120s -p 4 ./...

# Skip timeout for debugging
go test -timeout 0 -run TestSpecificTest ./internal/service
```

### Race detector errors

**Error:** `WARNING: DATA RACE` or `race condition detected`

**Solution:**

This indicates a genuine concurrency bug:

```bash
# First, run the test with race detector to see the error
go test -race -run TestFailing ./...

# Common causes:
# 1. Writing to shared map/slice without mutex
# 2. Not locking mutex before access
# 3. Goroutines accessing same variable

# Fix by:
# 1. Use sync.Mutex or sync.RWMutex
# 2. Always use the same lock for a resource
# 3. Prefer passing data via channels
```

### Test database connection failures

**Error:** `FAILED — unable to connect to test database`

**Solution:**

```bash
# Integration tests require Docker
docker compose -f docker-compose.test.yml up -d

# Verify containers are healthy
docker compose -f docker-compose.test.yml ps

# Check logs
docker compose -f docker-compose.test.yml logs

# Then run integration tests
go test -tags=integration ./...
```

### Mock not being called

**Error:** `Expected call not received as expected:` in test

**Solution:**

Check mock assertions:

```go
// ✓ Good: Verify mock was called
mockRepo.AssertCalled(t, "GetUser", mock.Anything, "123")

// ✗ Common error: Wrong argument type
mockRepo.On("GetUser", mock.Anything, "123").Return(...)  // Expects exact match
mockRepo.On("GetUser", mock.MatchedBy(func(id string) bool { return true }), mock.Anything).Return(...)

// ✓ Fix: Use MatchedBy for complex assertions
mockRepo.On("GetUser", mock.MatchedBy(func(ctx context.Context) bool {
    return ctx != nil
}), "123").Return(...)
```

---

## Deployment Issues

### Docker image build fails

**Error:** `failed to solve with frontend dockerfile.v0`

**Solution:**

```bash
# Clear Docker build cache
docker builder prune -a

# Rebuild
docker build -f deploy/Dockerfile -t go-service-template:latest .

# Or use make target
make docker-build
```

### Kubernetes pod not starting

**Error:** `CrashLoopBackOff` or `ImagePullBackOff`

**Solution:**

```bash
# Check pod status
kubectl describe pod <pod-name> -n production

# View pod logs
kubectl logs <pod-name> -n production

# Common issues:
# 1. Image not found → verify repositoryImage, tag
# 2. Missing env vars → check configMap, secrets
# 3. Database unavailable → check DATABASE_DSN, network

# Get into pod for debugging
kubectl exec -it <pod-name> -n production -- /bin/bash

# Or run a debug pod
kubectl run -it --image=busybox debug-pod --restart=Never -n production
```

### Database migrations fail during deployment

**Error:** Migration job pods failing

**Solution:**

```bash
# Check migration job logs
kubectl logs job/migrations -n production

# Rerun migrations manually
kubectl delete job migrations -n production
kubectl apply -f migration-job.yaml

# Or connect to database directly
kubectl run -it --image=postgres:16 pg-debug --restart=Never -n production -- \
  psql postgres://user:pass@postgres:5432/db -c "SELECT * FROM schema_migrations"
```

### Service not receiving traffic

**Error:** `Connection refused` or timeout when accessing service

**Solution:**

```bash
# Check service exists
kubectl get svc -n production

# Check endpoints
kubectl get endpoints -n production

# Check if pods are healthy
kubectl get pods -n production
kubectl describe pod <pod-name> -n production

# Check readiness probe
kubectl logs <pod-name> -n production | grep "health"

# Port-forward for testing
kubectl port-forward svc/go-service-template 8080:8080 -n production
curl http://localhost:8080/health/ready
```

### Helm deployment fails

**Error:** `error: template: ...: function "include" not available`

**Solution:**

```bash
# Verify Helm syntax
helm lint deploy/charts/service

# Check for syntax errors
helm template release deploy/charts/service -f deploy/charts/service/values.yaml

# Ensure values file exists
ls -la deploy/charts/service/values*.yaml

# Update dependencies
helm dependency update deploy/charts/service
```

---

## Performance Issues

### High memory usage

**Error:** Service using excessive memory; pod being OOM killed

**Solution:**

```bash
# Profile memory usage
go tool pprof http://localhost:6060/debug/pprof/heap

# Check cache settings
# Review cache.Set() TTLs — shorter TTLs = lower memory
# Implement cache eviction policies

# Database connection pooling
# Review APP_DATABASE_MAX_OPEN_CONNS in .env
# Too high = memory overhead; too low = connection waiting
```

### Slow database queries

**Error:** Service is slow; checking slow query logs

**Solution:**

```bash
# Enable slow query logging in PostgreSQL
ALTER SYSTEM SET log_min_duration_statement = 1000;  -- Log queries > 1 second
SELECT pg_reload_conf();

# View slow queries
SELECT query, calls, mean_time FROM pg_stat_statements ORDER BY mean_time DESC;

# Add database indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_orders_user_id ON orders(user_id);
```

### High CPU usage

**Error:** CPU usage at 100%; service unresponsive

**Solution:**

```bash
# Profile CPU usage
go tool pprof http://localhost:6060/debug/pprof/profile

# Look for:
# 1. Busy loops
# 2. Inefficient algorithms (N² where N¹ possible)
# 3. Unnecessary allocations

# Check for goroutine leaks
curl http://localhost:6060/debug/pprof/goroutine | wc -l
```

### Cache not helping performance

**Error:** Cache hit rate is low; not improving latency

**Solution:**

```bash
# Verify cache is being used
grep "s.cache.Get\|s.cache.Set" internal/service/*.go

# Check TTL values
# Too short = cache evicted before next request
# Too long = stale data

# Implement cache warming
// Pre-load frequently accessed data on startup

// Monitor cache hits/misses
redis-cli INFO stats | grep hits

// Review which queries are cached
// Cache expensive operations (aggregations, external APIs)
```

---

## Getting Help

### Before asking for help:

1. **Search existing issues** — Problem likely already solved
2. **Check logs** — Most errors are logged
3. **Run linter** — `make lint`
4. **Run tests** — `make test`
5. **Try in isolation** — Test function, not whole app
6. **Collect information**:
   - Error message (exact)
   - What were you doing?
   - OS and Go version (`go version`, `uname -a`)
   - Steps to reproduce

### If issue persists:

1. Check [VS_CODE_SETUP.md](VS_CODE_SETUP.md) for local setup help
2. Review [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) for common patterns
3. Read [ARCHITECTURE.md](ARCHITECTURE.md) for design decisions
4. Open a GitHub Issue with reproductions steps:

```markdown
## Issue Description
What is the problem?

## Steps to Reproduce
1. Run `make docker-run`
2. ...

## Expected Behavior
What should happen?

## Actual Behavior
What actually happened?

## Environment
- Go version: (run `go version`)
- OS: (macOS/Linux/Windows)
- error message: (exact output)

## Logs/Output
```
(paste relevant error messages or logs)
```
```

---

## Performance Debugging Checklist

- [ ] Check logs: `docker logs <container>`
- [ ] Monitor resources: `docker stats`
- [ ] Profile heap: `go tool pprof http://localhost:6060/debug/pprof/heap`
- [ ] Profile CPU: `go tool pprof http://localhost:6060/debug/pprof/profile`
- [ ] Check database: `docker exec postgres psql -U user -d db -c "SELECT * FROM pg_stat_statements"`
- [ ] Monitor cache: `redis-cli INFO stats`
- [ ] Review goroutines: `curl http://localhost:6060/debug/pprof/goroutine`

---

Still stuck? Check out these resources:

- [Go Debugging Guide](https://go.dev/doc/tutorial/get-started)
- [GORM Documentation](https://gorm.io/)
- [pgx Documentation](https://github.com/jackc/pgx)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Docker Documentation](https://docs.docker.com/)
