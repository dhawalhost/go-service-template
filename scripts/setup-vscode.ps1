$ErrorActionPreference = "Stop"

function Write-Log([string]$Message) {
    Write-Host "==> $Message"
}

function Write-Warn([string]$Message) {
    Write-Host "[warn] $Message"
}

function Test-Command([string]$Name) {
    return $null -ne (Get-Command $Name -ErrorAction SilentlyContinue)
}

function Install-WithWinget([string]$PackageId, [string]$Label) {
    if (-not (Test-Command "winget")) {
        Write-Warn "winget not found. Install $Label manually."
        return
    }

    Write-Log "Installing $Label with winget"
    winget install --id $PackageId --accept-package-agreements --accept-source-agreements --silent
}

Write-Log "Checking MCP runtime prerequisites"

if (-not (Test-Command "uv")) {
    Install-WithWinget "astral-sh.uv" "uv"
} else {
    Write-Log "uv already installed"
}

if (-not (Test-Command "node")) {
    Install-WithWinget "OpenJS.NodeJS.LTS" "Node.js LTS"
} else {
    Write-Log "node already installed"
}

if (Test-Command "uv") {
    Write-Log "uv version: $(uv --version)"
} else {
    Write-Warn "uv is not installed. fetch MCP server may fail to start."
}

if (Test-Command "npx") {
    Write-Log "npx version: $(npx --version)"
} else {
    Write-Warn "npx is not available. memory MCP server may fail to start."
}

if (-not (Test-Command "code")) {
    Write-Warn "VS Code CLI (code) not found in PATH."
    Write-Warn "Open VS Code and run: Command Palette -> Shell Command: Install 'code' command in PATH"
    exit 0
}

Write-Log "Installing recommended VS Code extensions"
$extensions = @(
    "golang.go",
    "github.copilot",
    "github.copilot-chat",
    "ms-azuretools.vscode-docker",
    "redhat.vscode-yaml"
)

$installed = code --list-extensions

foreach ($ext in $extensions) {
    if ($installed -contains $ext) {
        Write-Log "Extension already installed: $ext"
    } else {
        Write-Log "Installing extension: $ext"
        code --install-extension $ext
    }
}

Write-Log "Setup complete"
Write-Log "Reload VS Code window to ensure MCP servers are picked up from .vscode/mcp.json"
