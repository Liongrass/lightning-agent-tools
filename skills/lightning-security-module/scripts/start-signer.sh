#!/usr/bin/env bash
# Start the remote signer lnd node.
#
# Usage:
#   start-signer.sh                          # Default (mainnet, background)
#   start-signer.sh --network testnet        # Testnet
#   start-signer.sh --foreground             # Run in foreground

set -e

LNGET_SIGNER_DIR="${LNGET_SIGNER_DIR:-$HOME/.lnget/signer}"
LND_SIGNER_DIR="${LND_SIGNER_DIR:-$HOME/.lnd-signer}"
NETWORK="mainnet"
FOREGROUND=false
EXTRA_ARGS=""
RPC_PORT=10012
CONF_FILE="$LNGET_SIGNER_DIR/signer-lnd.conf"

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
        --foreground)
            FOREGROUND=true
            shift
            ;;
        --rpc-port)
            RPC_PORT="$2"
            shift 2
            ;;
        --extra-args)
            EXTRA_ARGS="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: start-signer.sh [options]"
            echo ""
            echo "Start the remote signer lnd node."
            echo ""
            echo "Options:"
            echo "  --network NETWORK    Bitcoin network (default: mainnet)"
            echo "  --lnddir DIR         Signer lnd data directory (default: ~/.lnd-signer)"
            echo "  --foreground         Run in foreground (default: background)"
            echo "  --rpc-port PORT      Signer RPC port (default: 10012)"
            echo "  --extra-args ARGS    Additional lnd arguments"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

# Verify lnd is installed.
if ! command -v lnd &>/dev/null; then
    echo "Error: lnd not found. Run install.sh first." >&2
    exit 1
fi

# Check if signer is already running by checking the RPC port.
if curl -sk "https://localhost:$RPC_PORT/v1/state" &>/dev/null 2>&1; then
    echo "Signer lnd is already running on port $RPC_PORT."
    echo "Use stop-signer.sh to stop it first."
    exit 1
fi

# Verify config exists.
if [ ! -f "$CONF_FILE" ]; then
    echo "Error: Signer config not found at $CONF_FILE" >&2
    echo "Run setup-signer.sh first." >&2
    exit 1
fi

echo "=== Starting Signer LND ==="
echo "Network:  $NETWORK"
echo "Data dir: $LND_SIGNER_DIR"
echo "Config:   $CONF_FILE"
echo "RPC port: $RPC_PORT"
echo ""

LOG_FILE="$LNGET_SIGNER_DIR/signer-lnd.log"

if [ "$FOREGROUND" = true ]; then
    exec lnd \
        --lnddir="$LND_SIGNER_DIR" \
        --configfile="$CONF_FILE" \
        $EXTRA_ARGS
else
    nohup lnd \
        --lnddir="$LND_SIGNER_DIR" \
        --configfile="$CONF_FILE" \
        $EXTRA_ARGS \
        > "$LOG_FILE" 2>&1 &
    SIGNER_PID=$!
    echo "Signer lnd started in background (PID: $SIGNER_PID)"
    echo "Log file: $LOG_FILE"
    echo ""

    # Wait briefly and verify it's running.
    sleep 2
    if kill -0 "$SIGNER_PID" 2>/dev/null; then
        echo "Signer lnd is running."
    else
        echo "Error: signer lnd exited immediately. Check $LOG_FILE" >&2
        tail -20 "$LOG_FILE" 2>/dev/null
        exit 1
    fi

    echo ""
    echo "The signer is ready to accept connections from watch-only nodes."
    echo "Watch-only nodes connect to this machine on port $RPC_PORT."
fi
