#!/usr/bin/env sh
set -e

# Change to the directory containing this script
cd "$(dirname "$0")"

# Detect compose engine
if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    COMPOSE="docker compose"
elif command -v podman >/dev/null 2>&1; then
    COMPOSE="podman compose"
else
    # If not docker/podman, maybe it's a native install
    if [ -f "./point-mcp" ] && [ -d ".git" ]; then
        echo "Updating Point MCP (Native)..."
        git pull
        go build -o point-mcp ./cmd/point-mcp
        if command -v systemctl >/dev/null 2>&1 && systemctl is-active --quiet point-mcp; then
            sudo systemctl restart point-mcp
        fi
        echo "Done!"
        exit 0
    fi
    echo "Error: neither docker nor podman found, and no native build environment detected." >&2
    exit 1
fi

echo "Updating Point MCP (Docker)..."

$COMPOSE pull
$COMPOSE up -d

echo "Done! Point MCP has been updated."
