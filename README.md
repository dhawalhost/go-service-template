# go-service-template

A production-ready **GitHub template repository** for spinning up new Go microservices in minutes. Built on top of [`github.com/dhawalhost/gokit`](https://github.com/dhawalhost/gokit) — a shared foundational library that provides HTTP routing, middleware, structured logging, database access, caching, observability, and more.

> **How to use**: Click the **"Use this template"** button on GitHub, clone your new repo, follow the [quick start](#quick-start) steps, and rename the placeholders marked with `RENAME_ME` comments.

## Documentation Index

Comprehensive guides for every aspect of developing with this template:

| Guide | Purpose |
|---|---|
| **[DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)** | 👨‍💻 Complete development workflow — setup, testing, debugging, common tasks |
| **[VS_CODE_SETUP.md](VS_CODE_SETUP.md)** | 🚀 AI-assisted development with VS Code, GitHub Copilot, and MCP integration |
| **[CONTRIBUTING.md](CONTRIBUTING.md)** | 🤝 Contribution guidelines — code standards, PR process, code review |
| **[ARCHITECTURE.md](ARCHITECTURE.md)** | 🏗️ Design patterns and architectural decisions — layering, DI, caching, database |
| **[TESTING.md](TESTING.md)** | ✅ Testing strategies — unit tests, integration tests, mocking, coverage |
| **[DEPLOYMENT.md](DEPLOYMENT.md)** | 🚢 Production deployment — Docker, Kubernetes, cloud platforms, migrations |
| **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** | 🔧 Common issues and solutions — setup, development, testing, deployment |

**New to the project?** Start with [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) for a comprehensive walkthrough.

---

## Production-Ready Features

This template includes everything needed for production microservices:

### 🏗️ Architecture
- **Layered architecture** (Handler → Service → Repository) for testability and maintainability
- **Dependency injection** for clean code and easy testing
- **Interface-based design** for swappable implementations
- **Cache-aside pattern** for high-performance reads
- **Optional tenancy support** that can be enabled for multi-tenant services or left off for single-tenant services

### 🔄 Data Access
- **GORM** for write operations (CREATE, UPDATE, DELETE with transactions)
- **pgx/v5** for optimized read queries
- **PostgreSQL migrations** with automatic execution on startup
- **Redis caching** with configurable TTLs

### 📡 Observability
- **Structured logging** with `zap` (automatic request ID tracking)
- **Prometheus metrics** (HTTP, database, Redis)
- **OpenTelemetry tracing** (optional)
- **Health check endpoints** (liveness & readiness probes)
- **Grafana dashboards** pre-configured

### 🔐 Security
- **JWT authentication** with configurable expiry
- **CORS middleware** to prevent cross-origin attacks
- **Rate limiting** to prevent abuse
- **Input validation** patterns throughout
- **SQL injection prevention** via parameterized queries

### 🧪 Testing
- **Table-driven tests** for comprehensive coverage
- **Mock patterns** for unit testing without dependencies
- **Integration test helpers** for database testing
- **Test Docker Compose** stack for CI/CD

### 📦 DevOps
- **GitHub Actions** for CI (lint, vet, test)
- **Automated Docker builds** and registry pushes
- **Helm charts** for Kubernetes deployment
- **HPA (Horizontal Pod Autoscaling)** configured
- **Multi-stage Docker builds** for minimal image size

### 🚀 Developer Experience
- **VS Code setup automation** (`make vscode-setup`)
- **GitHub Copilot integration** with MCP servers
- **Pre-commit hooks** for code quality
- **Makefile targets** for common tasks
- **Comprehensive documentation** (this guide + 6 additional guides)

---

```
HTTP Request
     │
     ▼
Global Middleware (RequestID, optional TenantID, Logger, Recovery, CORS, RateLimit, JWT)
     │
     ▼
Handler (internal/handler/) — HTTP parsing, validation, response writing
     │
     ▼
Service (internal/service/) — business logic, cache-aside pattern
     │
     ▼
Repository (internal/repository/)
  ├── postgres.go (GORM — CRUD operations)
  └── pgx.go (pgx/v5 — high-performance reads)
     │
     ▼
PostgreSQL + Redis
```

---

## What's Included

```
go-service-template/
├── go.mod                          # Module definition + gokit dependency
├── go.sum
├── README.md
├── .gitignore
├── .env.example                    # All environment variables with defaults
├── .golangci.yml                   # Lint configuration
├── Makefile                        # Developer workflow targets
├── cmd/
│   └── server/
│       └── main.go                 # ★ Canonical wiring reference (fully commented)
├── config/
│   └── config.go                   # Config struct embedding gokit base config
├── internal/
│   ├── handler/
│   │   ├── handler.go              # Handler struct + constructor
│   │   ├── routes.go               # chi sub-router with all routes
│   │   ├── example.go              # Full CRUD handler implementations
│   │   └── health.go               # Health route comment (mounted via gokit)
│   ├── service/
│   │   ├── service.go              # Service interface + implementation (cache-aside)
│   │   └── example.go              # Domain types: Example, ListParams, etc.
│   └── repository/
│       ├── repository.go           # Repository interface + GORM model
│       ├── postgres.go             # GORM implementation (all 5 CRUD methods)
│       └── pgx.go                  # pgx/v5 implementation (optimised List)
├── migrations/
│   ├── 000001_create_examples_table.up.sql
│   └── 000001_create_examples_table.down.sql
├── api/
│   └── openapi.yaml                # Full OpenAPI 3.0.3 specification
├── deploy/
│   ├── Dockerfile                  # Multi-stage, distroless final image
│   ├── docker-compose.yml          # App + Postgres + Redis + Prometheus + Grafana
│   ├── prometheus.yml              # Prometheus scrape config
│   └── charts/
│       └── service/                # Helm chart with HPA, Ingress, ConfigMap
├── .github/
│   └── workflows/
│       ├── ci.yml                  # Lint, vet, test on PRs and main
│       └── release.yml             # Build + push Docker image on version tags
└── scripts/
    ├── setup.sh                    # One-shot local dev setup
    └── migrate.sh                  # Run database migrations
```

---

## Prerequisites

- **Go 1.25+** — [Download](https://go.dev/dl/)
- **Docker** & **Docker Compose** — [Download](https://docs.docker.com/get-docker/)
- **golangci-lint** — `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **gokit** — cloned to `../gokit` (sibling directory) until published to pkg.go.dev

---

## Quick Start

```bash
# 1. Clone the template (or use "Use this template" on GitHub)
git clone https://github.com/dhawalhost/go-service-template my-service
cd my-service

# 2. Copy environment config
cp .env.example .env

# 3. Bootstrap VS Code + MCP prerequisites (uv/node) and recommended extensions
make vscode-setup

# 4. Start the full local stack (app + postgres + redis + prometheus + grafana)
make docker-run

# OR run locally (requires postgres + redis already running)
# make loads .env automatically when present
make run

# If running without make:
# set -a; source .env; set +a; go run ./cmd/server
```

**What `make vscode-setup` does:**
- ✓ Installs `uv` (Python package manager for MCP fetch server)
- ✓ Installs `Node.js` (for MCP memory server)
- ✓ Installs recommended VS Code extensions (Go, Copilot, Docker, YAML)
- ✓ Sets up pre-commit Git hooks for code quality

See **[VS_CODE_SETUP.md](VS_CODE_SETUP.md)** for detailed guidance on VS Code integration, GitHub Copilot setup, and using AI agents for development.

**Cross-platform setup:**
- macOS/Linux: `bash scripts/setup-vscode.sh`
- Windows (PowerShell): `powershell -ExecutionPolicy Bypass -File scripts/setup-vscode.ps1`

If `code` CLI is missing, run: `Cmd+Shift+P` → `Shell Command: Install 'code' command in PATH`.

**The service will be available at:**
- **API**: http://localhost:8080
- **Metrics**: http://localhost:8080/metrics
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin / admin)

---

## How to Use as a Template

After cloning, replace `example`/`Example` with your domain entity name using the following steps:

```bash
# 1. Rename the Go module
find . -type f -name "*.go" -exec sed -i 's|github.com/dhawalhost/go-service-template|github.com/your-org/your-service|g' {} +
sed -i 's|github.com/dhawalhost/go-service-template|github.com/your-org/your-service|g' go.mod

# 2. Rename the entity (Example → User, examples → users, etc.)
find . -type f \( -name "*.go" -o -name "*.sql" -o -name "*.yaml" \) \
  -exec sed -i 's/Example/User/g; s/example/user/g; s/examples/users/g' {} +

# 3. Rename source files
mv internal/handler/example.go internal/handler/user.go
mv internal/service/example.go internal/service/user.go

# 4. Update migrations
mv migrations/000001_create_examples_table.up.sql migrations/000001_create_users_table.up.sql
mv migrations/000001_create_examples_table.down.sql migrations/000001_create_users_table.down.sql

# 5. Tidy and verify
go mod tidy
go vet ./...
```

Look for `RENAME_ME` comments in the source files for all places that need customization.

---

## Configuration

All configuration is via environment variables (prefixed `APP_`) or a YAML config file. See `.env.example` for all defaults.

| Variable | Type | Default | Description |
|---|---|---|---|
| `APP_SERVER_ADDR` | string | `:8080` | HTTP listen address |
| `APP_SERVER_READ_TIMEOUT` | duration | `30s` | HTTP read timeout |
| `APP_SERVER_WRITE_TIMEOUT` | duration | `30s` | HTTP write timeout |
| `APP_SERVER_IDLE_TIMEOUT` | duration | `120s` | HTTP idle connection timeout |
| `APP_SERVER_SHUTDOWN_TIMEOUT` | duration | `30s` | Graceful shutdown timeout |
| `APP_DATABASE_DSN` | string | — | PostgreSQL connection string |
| `APP_DATABASE_MAX_OPEN_CONNS` | int | `25` | Max open DB connections |
| `APP_DATABASE_MAX_IDLE_CONNS` | int | `5` | Max idle DB connections |
| `APP_DATABASE_CONN_MAX_LIFETIME` | duration | `5m` | Max connection lifetime |
| `APP_DATABASE_MIGRATIONS_PATH` | string | `./migrations` | Path to SQL migrations |
| `APP_REDIS_ADDR` | string | `localhost:6379` | Redis address |
| `APP_REDIS_PASSWORD` | string | `` | Redis password |
| `APP_REDIS_DB` | int | `0` | Redis database number |
| `APP_JWT_SECRET` | string | — | JWT signing secret (**change in production!**) |
| `APP_JWT_EXPIRY` | duration | `24h` | JWT token expiry |
| `APP_JWT_ISSUER` | string | `go-service-template` | JWT issuer claim |
| `APP_LOG_LEVEL` | string | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `APP_LOG_DEVELOPMENT` | bool | `false` | Enable development log format |
| `APP_TELEMETRY_ENABLED` | bool | `false` | Enable OpenTelemetry tracing |
| `APP_TELEMETRY_OTLP_ENDPOINT` | string | `http://localhost:4318` | OTLP endpoint |
| `APP_TELEMETRY_SERVICE_NAME` | string | `go-service-template` | Service name for telemetry |

---

## Database Guide

### GORM vs pgx

| | GORM (`postgres.go`) | pgx (`pgx.go`) |
|---|---|---|
| **Use for** | CRUD: Create, Update, Delete, GetByID | High-throughput reads: List |
| **Pros** | ORM features, auto-migrations, relationships | Fastest possible query performance |
| **Cons** | Slightly higher overhead | Verbose, manual SQL |

The `postgresRepo` and `pgxRepo` both satisfy the same `Repository` interface. In `main.go`, `repository.NewPostgres(db.GORM)` is used by default. You can swap in `repository.NewPgx(db.Pool)` for the `List` endpoint if throughput demands it.

### Migration Workflow

```bash
# Run all pending migrations
make migrate-up

# Roll back the last migration
make migrate-down

# Or use the script directly
APP_DATABASE_DSN=postgres://... bash scripts/migrate.sh up
```

Migrations live in `./migrations/` and are run automatically on startup via `database.RunMigrations`.

---

## API

All endpoints require a `Authorization: Bearer <jwt>` header (except health endpoints).

### List examples
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/examples?page=1&page_size=20&search=foo"
```

### Get example
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/examples/550e8400-e29b-41d4-a716-446655440000"
```

### Create example
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Example","description":"An example resource"}' \
  "http://localhost:8080/api/v1/examples"
```

### Update example
```bash
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Name"}' \
  "http://localhost:8080/api/v1/examples/550e8400-e29b-41d4-a716-446655440000"
```

### Delete example
```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/examples/550e8400-e29b-41d4-a716-446655440000"
```

See `api/openapi.yaml` for the full OpenAPI 3.0.3 specification.

---

## Deployment

### Docker Compose (local / staging)

```bash
# Build and start everything
make docker-run

# Or manually
docker compose -f deploy/docker-compose.yml up --build
```

### Helm (Kubernetes)

```bash
# Install with default values
helm install my-service ./deploy/charts/service \
  --set image.repository=ghcr.io/your-org/my-service \
  --set image.tag=v1.0.0

# Install with custom values file
helm install my-service ./deploy/charts/service -f my-values.yaml

# Upgrade
helm upgrade my-service ./deploy/charts/service --set image.tag=v1.1.0
```

Key Helm values to override:

| Value | Description |
|---|---|
| `image.repository` | Docker image repository |
| `image.tag` | Docker image tag |
| `autoscaling.enabled` | Enable HPA (default: `true`) |
| `ingress.enabled` | Enable Ingress (default: `false`) |

**See [DEPLOYMENT.md](DEPLOYMENT.md) for comprehensive deployment guide covering Docker, Kubernetes, AWS, GCP, Azure, and production best practices.**
| `env.*` | Environment variables (mounted via ConfigMap) |

---

## CI/CD

### `ci.yml` — Runs on every PR and push to `main`
1. Check `go mod tidy` is clean
2. `go vet ./...`
3. `golangci-lint run ./...`
4. `go test ./... -race -coverprofile=coverage.out`
5. Upload coverage artifact

### `release.yml` — Runs on every `v*` tag push
1. Build Docker image (multi-stage, distroless)
2. Push to GitHub Container Registry (`ghcr.io`)
3. Tags: `v1.2.3`, `v1.2`, `latest`

To release:
```bash
git tag v1.0.0
git push origin v1.0.0
```

---

## Getting Help

### 🆘 I'm stuck!

Before opening an issue:

1. **Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md)** — Most common issues are documented with solutions
2. **Read [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)** — Comprehensive development guide with examples
3. **Try the AI agent** — In VS Code: `Cmd+Shift+I` → Ask Copilot with `@developer-go`

**Still stuck?** Open a GitHub Issue with:
- Exact error message
- Steps to reproduce
- Your OS and Go version (`go version`)
- Relevant logs or output

### 📚 Learning Paths

**I'm new to Go:**
1. [Go Tour](https://go.dev/tour/welcome/1) (interactive)
2. [Effective Go](https://go.dev/doc/effective_go) (best practices)
3. This template's [ARCHITECTURE.md](ARCHITECTURE.md) (patterns used here)

**I want to understand the template:**
1. Start with [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)
2. Review [ARCHITECTURE.md](ARCHITECTURE.md) for design patterns
3. Look at `internal/handler/example.go` for endpoint implementation pattern
4. Study `internal/service/example.go` for business logic pattern
5. Check `internal/repository/postgres.go` for data access pattern

**I'm ready to contribute:**
1. Read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines
2. Check [DEVELOPMENT_GUIDE.md](DEVELOPER_GUIDE.md) for development workflow
3. Review [TESTING.md](TESTING.md) for test requirements
4. See [ARCHITECTURE.md](ARCHITECTURE.md) for design decisions

**I'm deploying to production:**
1. Follow [DEPLOYMENT.md](DEPLOYMENT.md) for step-by-step guide
2. Review config in [Configuration](#configuration) section
3. Set up monitoring and alerting
4. Test rollback procedure (documented in DEPLOYMENT.md)

### 💡 Quick Reference

| Task | Command | Reference | 
|---|---|---|
| Run the service | `make run` | [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md#running-the-service) |
| Run tests | `make test` | [TESTING.md](TESTING.md#running-tests) |
| Code quality checks | `make fmt lint vet` | [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md#code-formatting--linting) |
| Format code | `make fmt` | [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md#code-formatting--linting) |
| Add a feature | See DEVELOPER_GUIDE | [CONTRIBUTING.md](CONTRIBUTING.md#3-implement-changes) |
| Deploy to Kubernetes | `helm install` | [DEPLOYMENT.md](DEPLOYMENT.md#helm-recommended) |
| Debug an issue | Check TROUBLESHOOTING.md | [TROUBLESHOOTING.md](TROUBLESHOOTING.md) |
| Setup VS Code | `make vscode-setup` | [VS_CODE_SETUP.md](VS_CODE_SETUP.md) |

---

This template is built on [`github.com/dhawalhost/gokit`](https://github.com/dhawalhost/gokit), a shared Go library providing:

- `server` — HTTP server with graceful shutdown
- `router` — chi v5 router factory + `Mount` helper
- `middleware` — RequestID, Logger, Recovery, CORS, JWT, RateLimit, SecureHeaders
- `logger` — zap structured logger with global accessor
- `config` — viper-based config loading with env var support
- `database` — GORM + pgxpool dual-layer database client
- `cache` — Redis cache interface via go-redis/v9
- `health` — Liveness/readiness handler with dependency registration
- `observability` — Prometheus metrics + OpenTelemetry tracing
- `errors` — Structured error types with HTTP status mapping
- `response` — Standardised JSON response helpers
- `validator` — go-playground/validator singleton
- `pagination` — Offset-based pagination parameter parsing
- `idgen` — UUID/ULID generation helpers
- `testutil` — Test helpers and mock factories

> **Note**: Until gokit is published to [pkg.go.dev](https://pkg.go.dev), the `go.mod` uses a `replace` directive pointing to `../gokit`. Clone both repos as siblings and remove the directive once published.
