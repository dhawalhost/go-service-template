# Contributing Guide

Thank you for contributing to **go-service-template**! This document outlines the process for proposing changes, ensuring consistency, and maintaining high code quality.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Process](#development-process)
4. [Coding Standards](#coding-standards)
5. [Testing Requirements](#testing-requirements)
6. [Commit & PR Guidelines](#commit--pr-guidelines)
7. [Code Review Process](#code-review-process)
8. [Documentation](#documentation)

---

## Code of Conduct

- Be respectful and inclusive
- Welcome all skill levels
- Critique ideas, not people
- Report unacceptable behavior to maintainers

---

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Git
- VS Code (recommended; see [VS_CODE_SETUP.md](VS_CODE_SETUP.md))

### Setup

```bash
# Clone the repository
git clone https://github.com/your-org/go-service-template.git
cd go-service-template

# Install development tools and VS Code setup
make vscode-setup

# Start local infrastructure
make docker-run
```

See [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) for detailed setup.

---

## Development Process

### 1. Create an Issue (or Pick an Existing One)

Before starting large work, open an issue to discuss:
- The problem you're solving
- Proposed approach
- Potential impact

**Issue template:**
```markdown
### Problem
What problem does this solve?

### Proposed Solution
How do you plan to solve it?

### Alternatives Considered
Any other approaches?

### Additional Context
Screenshots, error logs, related issues?
```

### 2. Create a Feature Branch

```bash
# Update main branch
git fetch origin
git checkout main
git pull origin main

# Create feature branch (use descriptive names)
git checkout -b feature/add-user-authentication
git checkout -b fix/race-condition-in-cache
git checkout -b docs/add-deployment-guide

# Push to origin to mark work-in-progress
git push -u origin feature/add-user-authentication
```

**Branch naming conventions:**
- `feature/*` — New functionality
- `fix/*` — Bug fixes
- `docs/*` — Documentation
- `refactor/*` — Code restructuring
- `perf/*` — Performance improvements
- `test/*` — Test improvements
- `ci/*` — CI/CD changes

### 3. Implement Changes

Follow the layered architecture:

1. **Update domain types** (`internal/service/`)
2. **Add repository methods** (`internal/repository/`)
3. **Implement business logic** (`internal/service/`)
4. **Add HTTP handlers** (`internal/handler/`)
5. **Update routes** (`internal/handler/routes.go`)
6. **Create/update OpenAPI spec** (`api/openapi.yaml`)
7. **Add database migrations** if needed (`migrations/`)
8. **Write tests** (`*_test.go` files)
9. **Update documentation** (README, guides)

### 4. Run Quality Checks

```bash
# Format code
make fmt

# Run linter
make lint

# Type check
make vet

# Run all tests
make test

# Check vulnerabilities
make vuln

# Or run everything
make fmt lint vet test vuln
```

**The pre-commit hook automatically runs these**, so commit will fail if they don't pass.

To skip (not recommended):
```bash
git commit --no-verify
```

---

## Coding Standards

### Go Code Style

**Follow standard Go conventions:**

1. **Formatting:** `gofmt` (automatic)
2. **Import organization:** stdlib → external → internal
3. **Naming:** See [DEVELOPER_GUIDE.md — Naming Conventions](DEVELOPER_GUIDE.md#naming-conventions)

### Error Handling

```go
// ✓ Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ✗ Avoid: Silent errors
_ = repository.DeleteOldRecords(ctx)

// ✓ Good: Log errors with context
if err != nil {
    h.log.Error("database error", zap.Error(err), zap.String("user_id", id))
    http.Error(w, "Internal server error", http.StatusInternalServerError)
    return
}
```

### Code Comments

```go
// ✓ Good: Explain the "why", not the "what"
// Cache user data for 5 minutes to reduce database load during peak hours.
// This is safe because user updates are rare.
s.cache.Set(ctx, fmt.Sprintf("user:%s", id), user, 5*time.Minute)

// ✗ Avoid: State what the code obviously does
// Increment the counter by 1
count++
```

### Concurrency

```go
// ✓ Good: Comment about goroutine management
// Launch async job to send email; errors are logged
go func() {
    if err := s.sendWelcomeEmail(ctx, user.Email); err != nil {
        s.log.Error("failed to send welcome email", zap.Error(err))
    }
}()

// ✗ Avoid: Silent goroutine panics
go someFunction()  // If this panics, the service crashes silently
```

### Dependencies

- Minimize external dependencies (larger attack surface, slower build)
- Only add dependencies that provide significant value
- Prefer standard library when reasonable
- Keep Go version requirement minimal (1.25+ is current standard)

---

## Testing Requirements

All code contributions must include tests. Untested code will not be merged.

### Test Structure

```go
// File: internal/service/user_test.go
func TestCreateUser(t *testing.T) {
    // Arrange
    repo := &mockRepository{}
    cache := &mockCache{}
    log := zap.NewNop()
    svc := New(repo, cache, log)

    // Act
    user, err := svc.CreateUser(context.Background(), &CreateUserRequest{
        Name:  "Alice",
        Email: "alice@example.com",
    })

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
    assert.Equal(t, "alice@example.com", user.Email)
}
```

### Test Coverage

```bash
# Generate coverage report
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# View in browser to find gaps
```

**Minimum coverage targets:**
- **Service layer:** 80%+ (business logic is critical)
- **Handler layer:** 70%+ (HTTP-specific code is simpler)
- **Repository layer:** Test with integration tests (see driver folder)

### Table-Driven Tests

For functions with multiple cases, use table-driven tests:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid", "user@example.com", false},
        {"no at sign", "user", true},
        {"empty", "", true},
        {"spaces", "user @example.com", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
            }
        })
    }
}
```

See [TESTING.md](TESTING.md) for comprehensive testing strategies.

---

## Commit & PR Guidelines

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type:** `feat`, `fix`, `test`, `docs`, `refactor`, `chore`, `perf`

**Scope:** Component affected (optional): `service`, `handler`, `repo`, `config`

**Subject:** Imperative, present tense, lowercase, no period

```bash
# ✓ Good
git commit -m "feat(handler): add user authentication endpoint"
git commit -m "fix(service): prevent cache stampede on cache miss"
git commit -m "test(service): add table-driven tests for validation"
git commit -m "docs: add deployment guide"

# ✗ Avoid
git commit -m "Added user auth"
git commit -m "Fixed a bug"
git commit -m "Updated stuff"
```

### Pull Request Guidelines

**PR Template:**
```markdown
## Description
Brief summary of changes.

## Related Issues
Closes #123

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation

## Changes Made
- Detailed list
- Of changes
- Included in this PR

## Testing
How was this tested?
- [ ] Unit tests added
- [ ] Integration tests added
- [ ] Tested locally with docker-compose
- [ ] Load tested if performance critical

## Checklist
- [ ] Code follows style guides (make lint)
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
- [ ] Security implications reviewed if applicable
```

### PR Title

Use the same format as commit messages:

```
feat(handler): add user authentication endpoint
fix(service): prevent cache stampede on cache miss
docs: add deployment guide
```

---

## Code Review Process

### Before Submitting PR

```bash
# Ensure all checks pass
make fmt lint vet test vuln

# Or use pre-commit hook
git commit  # Will auto-fail if checks don't pass

# Update against main branch
git fetch origin
git rebase origin/main

# Push
git push origin feature/my-feature
```

### PR Review Checklist

Reviewers will check:

1. **Correctness** — Does the code do what it's supposed to?
2. **Design** — Is the architecture sound? Any anti-patterns?
3. **Tests** — Adequate test coverage? Are tests meaningful?
4. **Style** — Follows Go conventions? Readable variable names?
5. **Performance** — Any obvious inefficiencies? N+1 queries?
6. **Security** — Input validation? SQL injection? Auth checks?
7. **Documentation** — Updated README, comments, OpenAPI spec?
8. **Backwards Compatibility** — Any breaking changes documented?

### Feedback Process

1. **Address all comments** — Reply to each review comment
2. **Push fixes** — `git commit` and `git push` again
3. **Re-request review** — GitHub allows re-requesting review
4. **Request approval** — Proceed once review is approved

### Approval & Merge

- Requires at least one approval from a maintainer
- All CI checks must pass (lint, test, security scan)
- All conversations must be resolved
- Maintainer merges using "Squash and merge" for cleaner history

---

## Documentation

### What to Document

- **Code changes:** Comments explaining "why", not "what"
- **New endpoints:** Update `api/openapi.yaml`
- **New config options:** Add to `.env.example` and `README.md`
- **New features:** Update relevant guides (DEVELOPER_GUIDE.md, etc.)

### OpenAPI Documentation

When adding HTTP endpoints, update `api/openapi.yaml`:

```yaml
/users:
  post:
    summary: Create a new user
    tags:
      - Users
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/CreateUserRequest'
    responses:
      '201':
        description: User created successfully
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      '400':
        description: Invalid input
```

### README Updates

- Any new feature should be mentioned in README.md
- Update the "What's Included" section if adding files
- Add configuration options to the Configuration table

---

## Questions?

- Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues
- Review [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) for technical reference
- Ask in GitHub Discussions or Issues
- Open a PR with your question as a draft

**Happy contributing!** 🎉
