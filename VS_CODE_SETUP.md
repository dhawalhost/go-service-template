# VS Code Setup Guide

This guide explains how to set up VS Code with **Model Context Protocol (MCP)** integration for enhanced AI-assisted development, automatic code quality checks, and intelligent tooling.

## Table of Contents

1. [Overview](#overview)
2. [Automated Setup](#automated-setup)
3. [MCP Servers](#mcp-servers)
4. [Recommended Extensions](#recommended-extensions)
5. [GitHub Copilot Setup](#github-copilot-setup)
6. [Troubleshooting](#troubleshooting)

---

## Overview

VS Code in this repository is pre-configured with:

- **MCP (Model Context Protocol) Integration** — Connect AI assistants to fetch data, search GitHub, and manage persistent memory
- **Recommended Extensions** — Go, Copilot, Docker, YAML support
- **Pre-commit Hooks** — Automatic formatting and linting before commits
- **Custom AI Agents** — Specialized agents in `.github/agents/` for different development tasks

This setup dramatically improves developer experience by:
- ✓ AI access to real-time GitHub issues and pull requests
- ✓ Persistent memory of your preferences and frequent tasks
- ✓ Automatic code formatting on save
- ✓ Specialized AI personas for architecture, testing, reviews, etc.

---

## Automated Setup

### First Time (Recommended)

```bash
# From repository root
make vscode-setup
```

This command:
1. ✓ Installs `uv` (Python package manager — needed for MCP `fetch` server)
2. ✓ Installs `Node.js` (needed for MCP `memory` server)
3. ✓ Installs all recommended VS Code extensions
4. ✓ Installs pre-commit Git hooks
5. ✓ Verifies the `code` CLI is available

**What gets installed:**

| Tool | Used For | Version |
|---|---|---|
| **uv** | Running MCP fetch server for HTTP requests | Latest |
| **Node.js** | Running MCP memory server for persistent context | Latest LTS |
| **golang.go** | Go language support | Latest |
| **github.copilot** | AI code generation and assistance | Latest |
| **github.copilot-chat** | AI chat in VS Code | Latest |
| **ms-azuretools.vscode-docker** | Docker/Docker Compose support | Latest |
| **redhat.vscode-yaml** | YAML validation and completion | Latest |

### Manual Setup (by OS)

If `make vscode-setup` fails or you prefer manual steps:

**macOS/Linux:**
```bash
bash scripts/setup-vscode.sh
```

**Windows (PowerShell):**
```powershell
powershell -ExecutionPolicy Bypass -File scripts/setup-vscode.ps1
```

---

## MCP Servers

MCP (Model Context Protocol) is a standard for connecting AI models to external tools and data sources. This repository configures three MCP servers in `.vscode/mcp.json`:

### 1. GitHub Server

**Purpose:** GitHub Copilot gets direct access to GitHub APIs

**Configured via:** Copilot's built-in GitHub integration (no setup needed)

**Capabilities:**
- Search issues and pull requests
- Get issue/PR details
- Create/update issues
- Post comments
- Access repository metadata

**Example usage in chat:**
```
@copilot Search for open issues labeled "bug"
@copilot What's the status of PR #42?
@copilot Create an issue: "Add integration tests"
```

### 2. Fetch Server

**Purpose:** Copilot can make HTTP requests to external APIs

**Configured via:** `mcp.json` entry with `uvx mcp-server-fetch`

**Prerequisites:** `uv` installed (done by `make vscode-setup`)

**Capabilities:**
- GET/POST/PATCH/DELETE HTTP requests
- Access any public API
- Retrieve documentation from URLs

**Example usage:**
```
@copilot Fetch status from https://api.github.com/repos/{owner}/{repo}
@copilot Get API docs from https://api.example.com/docs
```

### 3. Memory Server

**Purpose:** Copilot stores and retrieves persistent context about you and your preferences

**Configured via:** `mcp.json` entry with `npx @modelcontextprotocol/server-memory`

**Prerequisites:** `Node.js` installed (done by `make vscode-setup`)

**Capabilities:**
- Store user preferences ("I prefer test files named `*_test.go`")
- Remember common patterns in this codebase
- Track decisions and context between conversation sessions
- Persistent across VS Code restarts

**Example usage in chat:**
```
Remember that I always use table-driven tests for service logic
Remember: this team uses GORM for ORM and pgx for read optimization
Remember my Go style guide preferences
```

---

## Recommended Extensions

### 1. **Go** (golang.go)

Essential for Go development in VS Code.

**Features:**
- Syntax highlighting and IntelliSense
- Go to definition, Find all references
- Debugging with Delve
- Test runner
- Code generation (`go generate`, implement interfaces)

**Key Commands:**
- `Ctrl+Shift+D` — Open debugger
- `Ctrl+K Ctrl+X` — Organize imports
- `Cmd+Shift+P` → "Go: Add Tags to Struct Fields"

**Setup:**
```bash
# Already installed by make vscode-setup
# Manual install: search "Go" in Extensions marketplace
```

### 2. **GitHub Copilot** (github.copilot)

AI-powered code completion and suggestions.

**Features:**
- **Inline completions** — Smart code snippets as you type
- **Code explanation** — Understand unfamiliar code
- **Test generation** — Generate unit test scaffolds
- **Commit messages** — Auto-generate meaningful commits

**Setup:**
- Sign in with GitHub account (prompted on first use)
- Or manually: Click Copilot icon in status bar

**Key Usage:**
- `Tab` — Accept suggestion
- `Ctrl+]` — Dismiss suggestion
- `Option+]` — Next suggestion
- `Option+[` — Previous suggestion

### 3. **GitHub Copilot Chat** (github.copilot-chat)

AI chat panel for longer conversations and complex tasks.

**Features:**
- `@copilot` — Ask AI assistant
- `@workspace` — Include workspace files in context
- `#codebase` — Analyze specific files
- `/explain`, `/tests`, `/fix` — Specialized commands

**Opening Chat:**
- `Ctrl+Shift+I` (or `Cmd+Shift+I` on Mac) — Open chat panel
- `Ctrl+L` (in Editor) — Quick chat for current selection

### 4. **Docker** (ms-azuretools.vscode-docker)

Docker and Docker Compose support.

**Features:**
- Docker file syntax highlighting
- Docker Compose validation
- Run containers from explorer
- Build images interactively
- View running containers

**Key Commands:**
- `Cmd+Shift+P` → "Docker: Build Image"
- `Cmd+Shift+P` → "Docker Compose: Up"

### 5. **YAML** (redhat.vscode-yaml)

YAML validation and completion.

**Features:**
- YAML syntax validation
- Schema support (Kubernetes, Docker Compose, etc.)
- Auto-completion for known schemas
- Markdown-to-YAML code folding

**Configured for:**
- Kubernetes manifests in `deploy/charts/`
- Docker Compose validation in `deploy/docker-compose.yml`
- GitHub Actions workflow validation

---

## GitHub Copilot Setup

### Initial Authentication

1. **First use:** VS Code will prompt you to sign in with GitHub
2. **Manual sign-in:** Click the Copilot icon in the status bar → "Sign in to GitHub"
3. **Authenticate:** Browser opens → approve access → confirmation code

### Verify Setup

```bash
# In VS Code terminal:
1. Open any .go file
2. Press Ctrl+I (or Cmd+I on Mac)
3. Type: "How does this function work?"
4. Copilot should provide an explanation

# If it says "No authorization", re-authenticate via status bar
```

### Using Copilot Effectively

#### For Code Completion

```go
// Type this:
func Calculate

// Copilot suggests:
func CalculateSum(nums []int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// Review → Tab to accept, Esc to dismiss
```

#### For Code Review

```
Cmd+Shift+I (open chat)
@workspace
Ask: "Review this code for potential issues"

Copilot analyzes context and suggests improvements
```

#### For Test Generation

```go
// Place cursor in function
func CreateUser(ctx context.Context, name string) (*User, error) {
    // ...
}

// Press Ctrl+Shift+I
@copilot
Create unit tests for this function using table-driven tests
```

### Custom AI Agents

This repository includes specialized agents in `.github/agents/`:

- **developer-go.md** — Go backend implementation expertise
- **developer-react.md** — React/TypeScript frontend expertise
- **developer-devops.md** — Docker, Kubernetes, CI/CD
- **code-reviewer.md** — Code quality and security audit
- **tester.md** — Test strategy and coverage
- **technical-writer.md** — Documentation and API specs
- **architect.md** — System design and architecture

**Using Custom Agents:**

In Copilot Chat, you can reference these personas:

```
@developer-go
How should I structure error handling for this database operation?

@tester
What test cases am I missing for user authentication?

@architect
Is our current microservice split appropriate?
```

The agents have contextual knowledge specific to their domain and can provide more focused guidance.

---

## Pre-commit Hooks

The setup installs Git hooks that run before every commit:

### What Hooks Do

```bash
make hooks
# or manually
git config core.hooksPath .githooks
chmod +x .githooks/pre-commit
```

### The Pre-commit Hook

Automatically runs before each commit:

```bash
gofmt -w .         # Format code
go vet ./...       # Type check
golangci-lint run  # Style & bug detection
```

**If checks fail:**
- Fix the issues shown
- Run `git add .`
- `git commit` again

**To skip hooks (not recommended):**
```bash
git commit --no-verify
```

---

## Workspace Settings

VS Code settings specific to this repository are in `.vscode/settings.json`:

```json
{
  "go.lintOnSave": "package",
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--timeout=5m"],
  "[go]": {
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "golang.go",
    "editor.gofmt": true
  },
  "yaml.schemas": {
    "https://json.schemastore.org/docker-compose.json": "docker-compose*.yml"
  }
}
```

To add custom settings:
1. `Cmd+Shift+P` → "Preferences: Open Workspace Settings (JSON)"
2. Add your settings
3. They automatically apply to everyone else who clones the repo

---

## Common Workflows

### 1. Create a New Feature

```
Open Copilot Chat (Cmd+Shift+I)

@developer-go
I need to add a new endpoint POST /users.
What's the recommended approach?

Copilot provides:
- Domain model structure
- Repository method signature
- Service layer implementation
- Handler implementation
- Route registration pattern
```

### 2. Write Comprehensive Tests

```
Cmd+Shift+I
@tester
Generate test cases for the user creation function.
Use table-driven tests. Cover happy path, validation errors, and edge cases.

Copilot generates test scaffold covering all scenarios.
```

### 3. Security Review Before PR

```
Cmd+Shift+I
@code-reviewer
Review this code for security vulnerabilities, error handling, and performance issues.

@workspace (if analyzing multiple files)

Copilot identifies:
- SQL injection risks
- Unchecked errors
- Resource leaks
- Race conditions
```

### 4. Understand Unfamiliar Code

```
Select function/block of code
Cmd+I (Inline Chat)
Copilot explains what it does, why it's structured that way
```

### 5. Generate Documentation

```
Cmd+Shift+I
@technical-writer
Generate API documentation (OpenAPI/Swagger format) for these endpoints:
[paste handler code]

Copilot generates OpenAPI spec sections.
```

---

## Troubleshooting

### Copilot Not Responding

1. Check authentication: Click Copilot icon in status bar
2. Verify internet connection
3. Reload window: `Cmd+R` (or F5)
4. Restart VS Code

### MCP Servers Not Working

Check the status bar at bottom of VS Code:

```
If you see: ⚠️  MCP
Click it to see error details.

Common issues:
- uv not installed: Run `make vscode-setup` again
- Node.js not installed: Run `make vscode-setup` again
- Port conflict: Another app using port 2345

Resolution: Run `make vscode-setup --force`
```

### Go Extension Issues

```bash
# Reinstall Go tools
Cmd+Shift+P → "Go: Install/Update Tools"
Select all → Install

# Or manually
go install github.com/golangci/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/tools/gopls@latest
```

### Settings Not Applied

```bash
# Clear VS Code cache
rm -rf ~/.vscode/

# Restart VS Code
Cmd+Q then open VS Code again
```

### Workspace Trust

If VS Code asks "Do you trust the authors of this workspace?":
- Click "Trust" — this workspace has been verified
- This enables all extensions, including Copilot

---

## Next Steps

1. ✓ Run `make vscode-setup` (if not done)
2. ✓ Open the repository in VS Code
3. ✓ Authenticate with GitHub Copilot when prompted
4. ✓ Read [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) for development workflow
5. ✓ Check out the custom agents in `.github/agents/` for specialized help

**Start Here:**
```
Cmd+Shift+I
@developer-go
Hello! I'm new to this repository. What should I work on first?
```

Copilot provides an onboarding plan specific to this codebase.

---

## Resources

- [Model Context Protocol Spec](https://spec.modelcontextprotocol.io/)
- [VS Code Settings Reference](https://code.visualstudio.com/docs/getstarted/settings)
- [Go in VS Code](https://github.com/golang/vscode-go)
- [GitHub Copilot Docs](https://docs.github.com/en/copilot)
