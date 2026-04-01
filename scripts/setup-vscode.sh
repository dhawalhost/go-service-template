#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

log() {
  printf "==> %s\n" "$1"
}

warn() {
  printf "[warn] %s\n" "$1"
}

has_cmd() {
  command -v "$1" >/dev/null 2>&1
}

install_pkg() {
  local pkg="$1"

  if has_cmd brew; then
    log "Installing $pkg via Homebrew"
    brew install "$pkg"
    return 0
  fi

  if has_cmd apt-get; then
    log "Installing $pkg via apt-get"
    sudo apt-get update -y
    sudo apt-get install -y "$pkg"
    return 0
  fi

  if has_cmd dnf; then
    log "Installing $pkg via dnf"
    sudo dnf install -y "$pkg"
    return 0
  fi

  if has_cmd yum; then
    log "Installing $pkg via yum"
    sudo yum install -y "$pkg"
    return 0
  fi

  if has_cmd pacman; then
    log "Installing $pkg via pacman"
    sudo pacman -Sy --noconfirm "$pkg"
    return 0
  fi

  if has_cmd zypper; then
    log "Installing $pkg via zypper"
    sudo zypper --non-interactive install "$pkg"
    return 0
  fi

  return 1
}

install_uv_if_missing() {
  if has_cmd uv; then
    log "uv already installed"
    return
  fi

  case "$(uname -s)" in
    Darwin)
      if ! install_pkg "uv"; then
        warn "Could not install uv automatically. Install manually: https://docs.astral.sh/uv/getting-started/installation/"
      fi
      ;;
    Linux)
      if ! install_pkg "uv"; then
        warn "uv package not found in this distro package manager."
        warn "Install manually: curl -LsSf https://astral.sh/uv/install.sh | sh"
      fi
      ;;
    *)
      warn "Unsupported OS for this script. Use scripts/setup-vscode.ps1 on Windows."
      ;;
  esac
}

install_node_if_missing() {
  if has_cmd node; then
    log "node already installed"
    return
  fi

  case "$(uname -s)" in
    Darwin)
      if ! install_pkg "node"; then
        warn "Could not install node automatically. Install manually: https://nodejs.org/"
      fi
      ;;
    Linux)
      if ! install_pkg "nodejs"; then
        warn "Could not install nodejs automatically. Install manually: https://nodejs.org/"
      fi
      ;;
    *)
      warn "Unsupported OS for this script. Use scripts/setup-vscode.ps1 on Windows."
      ;;
  esac
}

log "Checking MCP runtime prerequisites"
install_uv_if_missing
install_node_if_missing

if has_cmd uv; then
  log "uv version: $(uv --version)"
else
  warn "uv is not installed. fetch MCP server may fail to start."
fi

if has_cmd npx; then
  log "npx version: $(npx --version)"
else
  warn "npx is not available. memory MCP server may fail to start."
fi

if ! has_cmd code; then
  warn "VS Code CLI (code) not found in PATH."
  warn "Open VS Code and run: Command Palette -> Shell Command: Install 'code' command in PATH"
  exit 0
fi

log "Installing recommended VS Code extensions"
extensions=(
  "golang.go"
  "github.copilot"
  "github.copilot-chat"
  "ms-azuretools.vscode-docker"
  "redhat.vscode-yaml"
)

for ext in "${extensions[@]}"; do
  if code --list-extensions | grep -qx "$ext"; then
    log "Extension already installed: $ext"
  else
    log "Installing extension: $ext"
    code --install-extension "$ext"
  fi
done

log "Setup complete"
log "Reload VS Code window to ensure MCP servers are picked up from .vscode/mcp.json"
