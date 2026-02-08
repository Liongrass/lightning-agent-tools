#!/usr/bin/env bash
# Stop the remote signer lnd node gracefully.
#
# Usage:
#   stop-signer.sh                    # Graceful stop via lncli
#   stop-signer.sh --force            # SIGTERM immediately

set -e

LND_SIGNER_DIR="${LND_SIGNER_DIR:-$HOME/.lnd-signer}"
NETWORK="${NETWORK:-mainnet}"
RPC_PORT=10012
FORCE=false

# Parse arguments.
while [[ $# -gt 0 ]]; do
    case $1 in
        --force)
            FORCE=true
            shift
            ;;
        --network)
            NETWORK="$2"
            shift 2
            ;;
        --rpc-port)
            RPC_PORT="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: stop-signer.sh [--force] [--network NETWORK] [--rpc-port PORT]"
            echo ""
            echo "Stop the remote signer lnd node."
            echo ""
            echo "Options:"
            echo "  --force              Send SIGTERM immediately"
            echo "  --network NETWORK    Bitcoin network (default: mainnet)"
            echo "  --rpc-port PORT      Signer RPC port (default: 10012)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

# Check if signer is running by probing the RPC port.
if ! curl -sk "https://localhost:$RPC_PORT/v1/state" &>/dev/null 2>&1; then
    echo "Signer lnd is not running (port $RPC_PORT not responding)."
    exit 0
fi

echo "Stopping signer lnd..."

if [ "$FORCE" = true ]; then
    # Find the process listening on the signer RPC port and kill it.
    SIGNER_PID=$(lsof -ti ":$RPC_PORT" 2>/dev/null | head -1 || true)
    if [ -n "$SIGNER_PID" ]; then
        kill "$SIGNER_PID"
        echo "Sent SIGTERM to PID $SIGNER_PID."
    else
        echo "Warning: Could not find process on port $RPC_PORT." >&2
        exit 1
    fi
else
    # Try graceful shutdown via lncli.
    if lncli --rpcserver="localhost:$RPC_PORT" \
        --lnddir="$LND_SIGNER_DIR" \
        --network="$NETWORK" \
        stop 2>/dev/null; then
        echo "Graceful shutdown initiated."
    else
        echo "lncli stop failed, finding process on port $RPC_PORT..."
        SIGNER_PID=$(lsof -ti ":$RPC_PORT" 2>/dev/null | head -1 || true)
        if [ -n "$SIGNER_PID" ]; then
            kill "$SIGNER_PID"
            echo "Sent SIGTERM to PID $SIGNER_PID."
        else
            echo "Warning: Could not find process to stop." >&2
            exit 1
        fi
    fi
fi

# Wait for process to exit.
echo "Waiting for signer lnd to exit..."
for i in {1..15}; do
    if ! curl -sk "https://localhost:$RPC_PORT/v1/state" &>/dev/null 2>&1; then
        echo "Signer lnd stopped."
        exit 0
    fi
    sleep 1
done

echo "Warning: signer lnd did not exit within 15 seconds." >&2
echo "Use --force or manually kill the process." >&2
exit 1
