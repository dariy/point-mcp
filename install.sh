#!/usr/bin/env bash
# Point MCP Server — Interactive Setup Wizard
# Usage: bash install.sh [--method=docker|native] [--non-interactive]
set -euo pipefail

# ── Constants ──────────────────────────────────────────────────────────────────
REPO="dariy/point-mcp"
GITHUB_API="https://api.github.com/repos/${REPO}/releases/latest"
RAW_BASE="https://raw.githubusercontent.com/${REPO}/main"

# ── Color output ───────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; BOLD='\033[1m'; NC='\033[0m'

say()  { echo -e "${BLUE}▶${NC}  $*"; }
ok()   { echo -e "${GREEN}✓${NC}  $*"; }
warn() { echo -e "${YELLOW}⚠${NC}  $*"; }
err()  { echo -e "${RED}✗${NC}  $*" >&2; }
die()  { err "$*"; exit 1; }
hr()   { echo -e "${BLUE}────────────────────────────────────────────${NC}"; }

# ask "Question" "default" → echoes the answer (default if user hits Enter)
ask() {
  local prompt="$1" default="$2" answer=""
  local display; [ -n "$default" ] && display="${prompt} [${default}]: " || display="${prompt}: "
  IFS= read -rp "$(echo -e "${BOLD}${display}${NC}")" answer </dev/tty || true
  echo "${answer:-${default}}"
}

# ask_yn "Question" "y|n" → echoes "y" or "n"
ask_yn() {
  local prompt="$1" default="${2:-y}" answer=""
  local hint; [ "$default" = "y" ] && hint="Y/n" || hint="y/N"
  IFS= read -rp "$(echo -e "${BOLD}${prompt}${NC} [${hint}]: ")" answer </dev/tty || true
  answer="${answer:-$default}"
  case "$answer" in [Yy]*) echo "y";; *) echo "n";; esac
}

# ── Banner ─────────────────────────────────────────────────────────────────────
show_banner() {
  echo ""
  echo -e "${BOLD}${BLUE}  ╔═══════════════════════════════════╗${NC}"
  echo -e "${BOLD}${BLUE}  ║    Point MCP Server — Installer   ║${NC}"
  echo -e "${BOLD}${BLUE}  ╚═══════════════════════════════════╝${NC}"
  echo ""
}

check_os() {
  if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    warn "This script is optimized for Linux. Continuing anyway..."
  fi
}

# ── CLI argument parsing ────────────────────────────────────────────────────────
METHOD_ARG=""
NON_INTERACTIVE=false
LOCAL_MODE=false

for arg in "$@"; do
  case "$arg" in
    --method=docker)      METHOD_ARG="docker" ;;
    --method=native)      METHOD_ARG="native" ;;
    --non-interactive)    NON_INTERACTIVE=true ;;
    --local)              LOCAL_MODE=true ;;
    --help|-h)
      echo "Usage: bash install.sh [--method=docker|native] [--non-interactive] [--local]"
      echo ""
      echo "  --method=docker     Install using Docker Compose (default)"
      echo "  --method=native     Install as native Linux binary"
      echo "  --non-interactive   Accept all defaults without prompting"
      echo "  --local             Test mode: build Docker image locally"
      exit 0
      ;;
    *) warn "Unknown argument: $arg" ;;
  esac
done

# Wrapper: in non-interactive mode, always returns the default
maybe_ask() {
  if [ "$NON_INTERACTIVE" = "true" ]; then echo "$2"; else ask "$1" "$2"; fi
}

# ── Install method selection ───────────────────────────────────────────────────
pick_install_method() {
  if [ -n "$METHOD_ARG" ]; then echo "$METHOD_ARG"; return; fi

  echo "" >&2
  echo -e "How would you like to install Point MCP?" >&2
  echo -e "  ${BOLD}1)${NC} Docker / Podman  ${GREEN}(recommended — easiest, safest)${NC}" >&2
  echo -e "  ${BOLD}2)${NC} Native Linux binary" >&2
  echo "" >&2
  local choice
  choice=$(maybe_ask "Choose [1/2]" "1")
  case "$choice" in
    2|native) echo "native" ;;
    *)        echo "docker" ;;
  esac
}

# ── Config collection ──────────────────────────────────────────────────────────
collect_config() {
  local method="$1"
  echo ""
  say "Configuration"
  echo ""

  POINT_BASE_URL=${POINT_BASE_URL:-$(maybe_ask "Point API Base URL (e.g. https://blog.example.com)" "")}
  while [ -z "$POINT_BASE_URL" ]; do
    if [ "$NON_INTERACTIVE" = "true" ]; then die "POINT_BASE_URL is required in non-interactive mode."; fi
    warn "POINT_BASE_URL is required."
    POINT_BASE_URL=$(maybe_ask "Point API Base URL" "")
  done

  POINT_API_KEY=${POINT_API_KEY:-$(maybe_ask "Point API Key" "")}
  while [ -z "$POINT_API_KEY" ]; do
    if [ "$NON_INTERACTIVE" = "true" ]; then die "POINT_API_KEY is required in non-interactive mode."; fi
    warn "POINT_API_KEY is required."
    POINT_API_KEY=$(maybe_ask "Point API Key" "")
  done

  if [ "$method" = "docker" ]; then
    TRANSPORT="http"
    DEPLOY_PORT=${DEPLOY_PORT:-$(maybe_ask "Host Port (DEPLOY_PORT)" "8000")}
    INSTALL_DIR=${INSTALL_DIR:-$(maybe_ask "Install directory" "$(pwd)")}
  else
    TRANSPORT=${TRANSPORT:-$(maybe_ask "Transport method (stdio/http)" "stdio")}
    if [ "$TRANSPORT" = "http" ]; then
      DEPLOY_PORT=${DEPLOY_PORT:-$(maybe_ask "HTTP Port" "8000")}
    else
      DEPLOY_PORT="8000"
    fi
    INSTALL_DIR=${INSTALL_DIR:-$(maybe_ask "Install directory" "/opt/point-mcp")}
  fi

  if [ "$TRANSPORT" = "http" ]; then
    MCP_AUTH_TOKENS=${MCP_AUTH_TOKENS:-$(maybe_ask "MCP Auth Tokens (comma separated, leave blank to generate)" "")}
    if [ -z "$MCP_AUTH_TOKENS" ]; then
      MCP_AUTH_TOKENS=$(openssl rand -hex 32)
      say "Generated random token: ${MCP_AUTH_TOKENS}"
    fi
    MCP_BASE_URL=${MCP_BASE_URL:-$(maybe_ask "MCP Base URL (optional, for OAuth)" "")}
    MCP_PASSWORD=${MCP_PASSWORD:-$(maybe_ask "MCP Password (optional, for OAuth)" "")}
  else
    MCP_AUTH_TOKENS=""
    MCP_BASE_URL=""
    MCP_PASSWORD=""
  fi
}

# ── Docker helpers ─────────────────────────────────────────────────────────────
detect_compose_engine() {
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    COMPOSE="docker compose"
  elif command -v podman >/dev/null 2>&1; then
    COMPOSE="podman compose"
  else
    COMPOSE=""
  fi
}

ensure_compose() {
  detect_compose_engine
  if [ -n "$COMPOSE" ]; then
    ok "Found: $COMPOSE"
    return
  fi
  die "Docker or Podman with compose support is required for the Docker install method."
}

write_env_file() {
  local env_path="$1"
  local method="$2"
  say "Writing ${env_path}..."

  local http_addr=":8000"
  if [ "$method" = "native" ] && [ "$TRANSPORT" = "http" ]; then
    http_addr=":${DEPLOY_PORT}"
  fi

  cat > "$env_path" <<EOF
# Point MCP Server configuration
POINT_BASE_URL=${POINT_BASE_URL}
POINT_API_KEY=${POINT_API_KEY}
EOF

  if [ "$method" = "native" ]; then
    cat >> "$env_path" <<EOF
POINT_MCP_TRANSPORT=${TRANSPORT}
POINT_MCP_HTTP_ADDR=${http_addr}
EOF
  fi

  if [ "$TRANSPORT" = "http" ]; then
    echo "MCP_AUTH_TOKENS=${MCP_AUTH_TOKENS}" >> "$env_path"
    [ -n "$MCP_BASE_URL" ] && echo "MCP_BASE_URL=${MCP_BASE_URL}" >> "$env_path"
    [ -n "$MCP_PASSWORD" ] && echo "MCP_PASSWORD=${MCP_PASSWORD}" >> "$env_path"
  fi
  ok ".env written"
}

install_via_docker() {
  ensure_compose

  say "Creating install directory: $INSTALL_DIR"
  mkdir -p "$INSTALL_DIR"

  local script_dir
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

  if [ "$LOCAL_MODE" = "true" ]; then
    say "Local mode: copying docker-compose.test.yml and Dockerfile"
    cp "${script_dir}/docker-compose.test.yml" "${INSTALL_DIR}/docker-compose.yml"
    cp "${script_dir}/Dockerfile" "${INSTALL_DIR}/Dockerfile"
    cp "${script_dir}/go.mod" "${INSTALL_DIR}/go.mod"
    cp "${script_dir}/go.sum" "${INSTALL_DIR}/go.sum"
    cp -r "${script_dir}/cmd" "${INSTALL_DIR}/cmd"
    cp -r "${script_dir}/internal" "${INSTALL_DIR}/internal"
  else
    if [ -f "${script_dir}/docker-compose.yml" ]; then
      # We strip the "build: ." line to prevent requiring the Dockerfile in the install dir
      grep -v "build: ." "${script_dir}/docker-compose.yml" > "${INSTALL_DIR}/docker-compose.yml"
    else
      curl -fsSL "${RAW_BASE}/docker-compose.yml" | grep -v "build: ." > "${INSTALL_DIR}/docker-compose.yml"
    fi
  fi

  if [ -f "${script_dir}/update.sh" ]; then
    cp "${script_dir}/update.sh" "${INSTALL_DIR}/update.sh"
    chmod +x "${INSTALL_DIR}/update.sh"
  fi

  # Write configuration
  write_env_file "${INSTALL_DIR}/.env" "docker"
  echo "DEPLOY_PORT=${DEPLOY_PORT}" >> "${INSTALL_DIR}/.env"

  say "Starting Point MCP..."
  (cd "$INSTALL_DIR" && $COMPOSE up -d)
  ok "Container started"
}

# ── Native install helpers ─────────────────────────────────────────────────────
install_native() {
  local script_dir
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  
  say "Building Point MCP..."
  if ! command -v go >/dev/null 2>&1; then
    die "Go compiler not found. Please install Go (1.26+) or use Docker."
  fi

  mkdir -p "$INSTALL_DIR"
  (cd "$script_dir" && go build -o "${INSTALL_DIR}/point-mcp" ./cmd/point-mcp)
  ok "Binary built to ${INSTALL_DIR}/point-mcp"

  if [ -f "${script_dir}/update.sh" ]; then
    cp "${script_dir}/update.sh" "${INSTALL_DIR}/update.sh"
    chmod +x "${INSTALL_DIR}/update.sh"
  fi

  write_env_file "${INSTALL_DIR}/.env" "native"

  if [ "$TRANSPORT" = "http" ] && [ "$(id -u)" -eq 0 ]; then
    local setup_svc
    setup_svc=$(ask_yn "Install systemd service?" "y")
    if [ "$setup_svc" = "y" ]; then
      say "Installing systemd service..."
      cat > /etc/systemd/system/point-mcp.service <<EOF
[Unit]
Description=Point MCP Server
After=network.target

[Service]
Type=simple
WorkingDirectory=${INSTALL_DIR}
EnvironmentFile=${INSTALL_DIR}/.env
ExecStart=${INSTALL_DIR}/point-mcp
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
      systemctl daemon-reload
      systemctl enable --now point-mcp
      ok "Service enabled and started"
    fi
  fi
}

show_success() {
  echo ""
  hr
  echo -e "${GREEN}${BOLD}  Point MCP Server is ready!${NC}"
  hr
  echo ""
  if [ "$TRANSPORT" = "http" ]; then
    echo -e "  ${BOLD}HTTP Address:${NC} http://localhost:${DEPLOY_PORT}"
    echo -e "  ${BOLD}Auth Token:${NC}   ${MCP_AUTH_TOKENS}"
  else
    echo -e "  ${BOLD}Transport:${NC}    stdio"
    echo -e "  Run the server: ${INSTALL_DIR}/point-mcp"
  fi
  echo ""
}

# ── Main ───────────────────────────────────────────────────────────────────────
main() {
  show_banner
  check_os

  INSTALL_METHOD=$(pick_install_method)
  collect_config "$INSTALL_METHOD"

  hr
  say "Starting installation (method: ${BOLD}${INSTALL_METHOD}${NC})"
  hr

  if [ "$INSTALL_METHOD" = "docker" ]; then
    install_via_docker
  else
    install_native
  fi

  show_success
}

main "$@"
