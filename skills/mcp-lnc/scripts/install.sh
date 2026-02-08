#!/usr/bin/env bash
# Build and install the MCP LNC server from source.
#
# Usage:
#   install.sh              # Build from mcp-server/ in this repo
#   install.sh --release    # Build optimized release binary
#
# Prerequisites: Go 1.24+

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MCP_SERVER_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)/mcp-server"
RELEASE=false

# Parse arguments.
while [[ $# -gt 0 ]]; do
    case $1 in
        --release)
            RELEASE=true
            shift
            ;;
        -h|--help)
            echo "Usage: install.sh [--release]"
            echo ""
            echo "Build and install the MCP LNC server."
            echo ""
            echo "Options:"
            echo "  --release  Build optimized release binary (no dev tags)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

echo "=== Installing MCP LNC Server ==="
echo ""

# Verify Go is installed.
if ! command -v go &>/dev/null; then
    echo "Error: Go is not installed." >&2
    echo "Install Go from https://go.dev/dl/" >&2
    exit 1
fi

GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | head -1)
echo "Go version: $GO_VERSION"

# Verify mcp-server directory exists.
if [ ! -d "$MCP_SERVER_DIR" ]; then
    echo "Error: mcp-server/ directory not found at $MCP_SERVER_DIR" >&2
    echo "Are you running this from the lightning-agent-kit repo?" >&2
    exit 1
fi

if [ ! -f "$MCP_SERVER_DIR/go.mod" ]; then
    echo "Error: go.mod not found in $MCP_SERVER_DIR" >&2
    exit 1
fi

echo "Source:  $MCP_SERVER_DIR"
echo ""

GOBIN=$(go env GOPATH)/bin

# Build the binary.
cd "$MCP_SERVER_DIR"

if [ "$RELEASE" = true ]; then
    echo "Building release binary..."
    COMMIT=$(git describe --abbrev=40 --always --dirty 2>/dev/null || echo "unknown")
    PKG="github.com/lightninglabs/lightning-agent-kit/mcp-server"
    go build -v -ldflags "-X ${PKG}.Commit=${COMMIT}" -o "$GOBIN/mcp-lnc-server" .
else
    echo "Building debug binary..."
    COMMIT=$(git describe --abbrev=40 --always --dirty 2>/dev/null || echo "unknown")
    PKG="github.com/lightninglabs/lightning-agent-kit/mcp-server"
    go build -v -tags "dev" -ldflags "-X ${PKG}.Commit=${COMMIT}" -o "$GOBIN/mcp-lnc-server" .
fi

echo "Done."
echo ""

# Verify installation.
if command -v mcp-lnc-server &>/dev/null; then
    echo "mcp-lnc-server installed: $(which mcp-lnc-server)"
    mcp-lnc-server -version 2>/dev/null || true
else
    echo "Warning: mcp-lnc-server not found on PATH." >&2
    echo "Binary built at: $GOBIN/mcp-lnc-server" >&2
    echo "Ensure \$GOPATH/bin is in your PATH." >&2
    echo "  export PATH=\$PATH:\$(go env GOPATH)/bin" >&2
fi

echo ""
echo "Installation complete."
echo ""
echo "Next steps:"
echo "  1. Configure: skills/mcp-lnc/scripts/configure.sh"
echo "  2. Add to Claude Code: skills/mcp-lnc/scripts/setup-claude-config.sh"
