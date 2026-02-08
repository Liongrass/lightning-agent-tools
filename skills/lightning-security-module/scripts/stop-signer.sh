#!/usr/bin/env bash
# Stop the remote signer lnd node gracefully.
#
# Usage:
#   stop-signer.sh                    # Graceful stop via lncli
#   stop-signer.sh --container sam    # Stop signer in Docker container
#   stop-signer.sh --rpcserver remote:10012 --tlscertpath ~/tls.cert --macaroonpath ~/admin.macaroon
#   stop-signer.sh --force            # SIGTERM immediately

set -e

LND_SIGNER_DIR="${LND_SIGNER_DIR:-}"
NETWORK="${NETWORK:-mainnet}"
RPC_PORT=10012
FORCE=false
CONTAINER=""
RPCSERVER=""
TLSCERTPATH=""
MACAROONPATH=""

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
        --container)
            CONTAINER="$2"
            shift 2
            ;;
        --rpcserver)
            RPCSERVER="$2"
            shift 2
            ;;
        --tlscertpath)
            TLSCERTPATH="$2"
            shift 2
            ;;
        --macaroonpath)
            MACAROONPATH="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: stop-signer.sh [options]"
            echo ""
            echo "Stop the remote signer lnd node."
            echo ""
            echo "Options:"
            echo "  --force                Send SIGTERM immediately (or docker stop for containers)"
            echo "  --network NETWORK      Bitcoin network (default: mainnet)"
            echo "  --rpc-port PORT        Signer RPC port (default: 10012)"
            echo "  --container NAME       Stop lnd running inside a Docker container"
            echo "  --rpcserver HOST:PORT  Connect to a remote signer node"
            echo "  --tlscertpath PATH     TLS certificate for remote connection"
            echo "  --macaroonpath PATH    Macaroon for remote authentication"
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

if [ -n "$CONTAINER" ]; then
    # Docker container mode.
    if ! docker ps --format '{{.Names}}' | grep -qx "$CONTAINER"; then
        echo "Container '$CONTAINER' is not running."
        exit 0
    fi

    echo "Stopping signer lnd in container '$CONTAINER'..."

    if [ "$FORCE" = true ]; then
        docker stop "$CONTAINER"
        echo "Container stopped."
    else
        if docker exec "$CONTAINER" lncli --rpcserver="localhost:$RPC_PORT" \
            --lnddir="$LND_SIGNER_DIR" \
            --network="$NETWORK" \
            stop 2>/dev/null; then
            echo "Graceful shutdown initiated."
        else
            echo "lncli stop failed, stopping container..."
            docker stop "$CONTAINER"
            echo "Container stopped."
        fi
    fi
    exit 0
fi

# Build connection flags for lncli.
CONN_FLAGS=(--network="$NETWORK" --lnddir="$LND_SIGNER_DIR")
if [ -n "$RPCSERVER" ]; then
    CONN_FLAGS+=("--rpcserver=$RPCSERVER")
else
    CONN_FLAGS+=("--rpcserver=localhost:$RPC_PORT")
fi
if [ -n "$TLSCERTPATH" ]; then
    CONN_FLAGS+=("--tlscertpath=$TLSCERTPATH")
fi
if [ -n "$MACAROONPATH" ]; then
    CONN_FLAGS+=("--macaroonpath=$MACAROONPATH")
fi

# Remote mode — stop via lncli only (no PID or port access).
if [ -n "$RPCSERVER" ]; then
    echo "Stopping remote signer at $RPCSERVER..."
    if lncli "${CONN_FLAGS[@]}" stop; then
        echo "Graceful shutdown initiated."
    else
        echo "Error: lncli stop failed for remote signer." >&2
        exit 1
    fi
    exit 0
fi

# Local mode — check if signer is running by probing the RPC port.
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
    if lncli "${CONN_FLAGS[@]}" stop 2>/dev/null; then
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
