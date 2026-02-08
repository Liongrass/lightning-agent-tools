#!/usr/bin/env bash
# Install lnd and lncli on the signer machine.
#
# Usage:
#   install.sh              # Install latest release
#   install.sh --version v0.19.2-beta  # Specific version
#
# Prerequisites: Go 1.21+
#
# Note: lnd uses replace directives in go.mod, so `go install` from the
# module registry does not work. This script clones the repo and builds.

set -e

VERSION=""
BUILD_TAGS="signrpc walletrpc chainrpc invoicesrpc routerrpc peersrpc kvdb_sqlite neutrinorpc"

# Parse arguments.
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --tags)
            BUILD_TAGS="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: install.sh [--version VERSION] [--tags TAGS]"
            echo ""
            echo "Install lnd and lncli on the signer machine."
            echo ""
            echo "Options:"
            echo "  --version VERSION  Git tag (e.g., v0.19.2-beta). Default: latest release."
            echo "  --tags TAGS        Build tags (default: signrpc walletrpc ...)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

echo "=== Installing lnd (signer) ==="
echo ""

# Verify Go is installed.
if ! command -v go &>/dev/null; then
    echo "Error: Go is not installed." >&2
    echo "Install Go from https://go.dev/dl/" >&2
    exit 1
fi

GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | head -1)
echo "Go version: $GO_VERSION"
echo "Build tags: $BUILD_TAGS"
echo ""

# Clone lnd into a temp directory and build from source.
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

echo "Cloning lnd..."
git clone --quiet https://github.com/lightningnetwork/lnd.git "$TMPDIR/lnd"

cd "$TMPDIR/lnd"

# Checkout specific version if requested, otherwise use latest tag.
if [ -n "$VERSION" ]; then
    echo "Checking out $VERSION..."
    git checkout --quiet "$VERSION"
else
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -n "$LATEST_TAG" ]; then
        echo "Using latest tag: $LATEST_TAG"
        git checkout --quiet "$LATEST_TAG"
    else
        echo "Using HEAD (no tags found)."
    fi
fi
echo ""

GOBIN=$(go env GOPATH)/bin

# Build lnd.
echo "Building lnd..."
go build -tags "$BUILD_TAGS" -o "$GOBIN/lnd" ./cmd/lnd
echo "Done."

# Build lncli.
echo "Building lncli..."
go build -tags "$BUILD_TAGS" -o "$GOBIN/lncli" ./cmd/lncli
echo "Done."
echo ""

# Verify installation.
if command -v lnd &>/dev/null; then
    echo "lnd installed: $(which lnd)"
    lnd --version 2>/dev/null || true
else
    echo "Warning: lnd not found on PATH." >&2
    echo "Ensure \$GOPATH/bin is in your PATH." >&2
    echo "  export PATH=\$PATH:\$(go env GOPATH)/bin" >&2
fi

if command -v lncli &>/dev/null; then
    echo "lncli installed: $(which lncli)"
else
    echo "Warning: lncli not found on PATH." >&2
fi

echo ""
echo "Installation complete."
echo ""
echo "Next steps:"
echo "  1. Set up signer: skills/lightning-security-module/scripts/setup-signer.sh"
