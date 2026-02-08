#!/usr/bin/env bash
# Export credentials bundle from a running signer for watch-only node import.
#
# Usage:
#   export-credentials.sh                               # Default (local signer)
#   export-credentials.sh --container sam               # Signer in Docker
#   export-credentials.sh --network testnet             # Testnet
#   export-credentials.sh --output /path/to/output      # Custom output dir
#
# Produces:
#   ~/.lnget/signer/credentials-bundle/accounts.json
#   ~/.lnget/signer/credentials-bundle/tls.cert
#   ~/.lnget/signer/credentials-bundle/admin.macaroon
#   ~/.lnget/signer/credentials-bundle.tar.gz.b64       (portable base64)

set -e

LNGET_SIGNER_DIR="${LNGET_SIGNER_DIR:-$HOME/.lnget/signer}"
LND_SIGNER_DIR="${LND_SIGNER_DIR:-}"
NETWORK="mainnet"
RPC_PORT=10012
OUTPUT_DIR=""
CONTAINER=""

# Parse arguments.
while [[ $# -gt 0 ]]; do
    case $1 in
        --network)
            NETWORK="$2"
            shift 2
            ;;
        --lnddir)
            LND_SIGNER_DIR="$2"
            shift 2
            ;;
        --rpc-port)
            RPC_PORT="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --container)
            CONTAINER="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: export-credentials.sh [options]"
            echo ""
            echo "Export credentials bundle from a running signer."
            echo ""
            echo "Options:"
            echo "  --network NETWORK   Bitcoin network (default: mainnet)"
            echo "  --lnddir DIR        Signer lnd data directory (default: ~/.lnd-signer)"
            echo "  --rpc-port PORT     Signer RPC port (default: 10012)"
            echo "  --output DIR        Output directory (default: ~/.lnget/signer/credentials-bundle)"
            echo "  --container NAME    Export from lnd running inside a Docker container"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

# Apply default lnddir if not set.
if [ -z "$LND_SIGNER_DIR" ]; then
    if [ -n "$CONTAINER" ]; then
        LND_SIGNER_DIR="/root/.lnd"
    else
        LND_SIGNER_DIR="$HOME/.lnd-signer"
    fi
fi

BUNDLE_DIR="${OUTPUT_DIR:-$LNGET_SIGNER_DIR/credentials-bundle}"

echo "=== Exporting Credentials Bundle ==="
echo ""
echo "Network:    $NETWORK"
echo "Signer dir: $LND_SIGNER_DIR"
echo "Output:     $BUNDLE_DIR"
echo ""

# Verify lncli is available.
if [ -n "$CONTAINER" ]; then
    if ! docker exec "$CONTAINER" which lncli &>/dev/null; then
        echo "Error: lncli not found in container '$CONTAINER'." >&2
        exit 1
    fi
elif ! command -v lncli &>/dev/null; then
    echo "Error: lncli not found. Run install.sh first." >&2
    exit 1
fi

# Create bundle directory.
mkdir -p "$BUNDLE_DIR"
chmod 700 "$BUNDLE_DIR"

# Export accounts list.
echo "Exporting accounts..."
if [ -n "$CONTAINER" ]; then
    docker exec "$CONTAINER" lncli --rpcserver="localhost:$RPC_PORT" \
        --lnddir="$LND_SIGNER_DIR" \
        --network="$NETWORK" \
        wallet accounts list > "$BUNDLE_DIR/accounts.json"
else
    lncli --rpcserver="localhost:$RPC_PORT" \
        --lnddir="$LND_SIGNER_DIR" \
        --network="$NETWORK" \
        wallet accounts list > "$BUNDLE_DIR/accounts.json"
fi

if [ ! -s "$BUNDLE_DIR/accounts.json" ]; then
    echo "Error: Failed to export accounts. Is the signer running and unlocked?" >&2
    exit 1
fi
echo "  accounts.json exported."

# Copy TLS certificate.
TLS_CERT="$LND_SIGNER_DIR/tls.cert"
if [ -n "$CONTAINER" ]; then
    docker cp "$CONTAINER:$TLS_CERT" "$BUNDLE_DIR/tls.cert" 2>/dev/null
    if [ ! -f "$BUNDLE_DIR/tls.cert" ]; then
        echo "Error: TLS certificate not found at $TLS_CERT in container '$CONTAINER'" >&2
        exit 1
    fi
else
    if [ ! -f "$TLS_CERT" ]; then
        echo "Error: TLS certificate not found at $TLS_CERT" >&2
        exit 1
    fi
    cp "$TLS_CERT" "$BUNDLE_DIR/tls.cert"
fi
echo "  tls.cert copied."

# Copy admin macaroon.
MACAROON="$LND_SIGNER_DIR/data/chain/bitcoin/$NETWORK/admin.macaroon"
if [ -n "$CONTAINER" ]; then
    docker cp "$CONTAINER:$MACAROON" "$BUNDLE_DIR/admin.macaroon" 2>/dev/null
    if [ ! -f "$BUNDLE_DIR/admin.macaroon" ]; then
        echo "Error: Admin macaroon not found at $MACAROON in container '$CONTAINER'" >&2
        exit 1
    fi
else
    if [ ! -f "$MACAROON" ]; then
        echo "Error: Admin macaroon not found at $MACAROON" >&2
        exit 1
    fi
    cp "$MACAROON" "$BUNDLE_DIR/admin.macaroon"
fi
echo "  admin.macaroon copied."
echo ""

# Create portable base64-encoded tar.gz bundle.
BUNDLE_ARCHIVE="$LNGET_SIGNER_DIR/credentials-bundle.tar.gz.b64"
echo "Creating portable bundle..."
tar -czf - -C "$BUNDLE_DIR" accounts.json tls.cert admin.macaroon | base64 > "$BUNDLE_ARCHIVE"
echo "  Bundle saved to $BUNDLE_ARCHIVE"
echo ""

echo "=== Credentials Bundle Ready ==="
echo ""
echo "Bundle contents:"
echo "  $BUNDLE_DIR/accounts.json    — account xpubs for watch-only import"
echo "  $BUNDLE_DIR/tls.cert         — signer TLS certificate"
echo "  $BUNDLE_DIR/admin.macaroon   — signer admin macaroon"
echo ""
echo "Portable bundle (base64):"
echo "  $BUNDLE_ARCHIVE"
echo ""
echo "To transfer, either:"
echo "  1. Copy the credentials-bundle/ directory to the agent machine"
echo "  2. Copy-paste the base64 string from $BUNDLE_ARCHIVE"
echo ""
echo "On the agent machine:"
echo "  skills/lnd/scripts/import-credentials.sh --bundle <path-or-base64>"
