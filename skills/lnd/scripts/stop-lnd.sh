#!/usr/bin/env bash
# Stop lnd gracefully.
#
# Usage:
#   stop-lnd.sh                    # Graceful stop via lncli
#   stop-lnd.sh --container sam    # Stop lnd in Docker container
#   stop-lnd.sh --force            # SIGTERM immediately

set -e

LND_DIR="${LND_DIR:-}"
NETWORK="${NETWORK:-mainnet}"
FORCE=false
CONTAINER=""

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
        --container)
            CONTAINER="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: stop-lnd.sh [--force] [--container NAME] [--network NETWORK]"
            echo ""
            echo "Stop lnd gracefully."
            echo ""
            echo "Options:"
            echo "  --force              Send SIGTERM immediately (or docker stop for containers)"
            echo "  --network NETWORK    Bitcoin network (default: mainnet)"
            echo "  --container NAME     Stop lnd running inside a Docker container"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

# Apply default lnddir if not set.
if [ -z "$LND_DIR" ]; then
    if [ -n "$CONTAINER" ]; then
        LND_DIR="/root/.lnd"
    else
        LND_DIR="$HOME/.lnd"
    fi
fi

if [ -n "$CONTAINER" ]; then
    # Docker container mode.
    if ! docker ps --format '{{.Names}}' | grep -qx "$CONTAINER"; then
        echo "Container '$CONTAINER' is not running."
        exit 0
    fi

    echo "Stopping lnd in container '$CONTAINER'..."

    if [ "$FORCE" = true ]; then
        docker stop "$CONTAINER"
        echo "Container stopped."
    else
        if docker exec "$CONTAINER" lncli --network="$NETWORK" --lnddir="$LND_DIR" stop 2>/dev/null; then
            echo "Graceful shutdown initiated."
        else
            echo "lncli stop failed, stopping container..."
            docker stop "$CONTAINER"
            echo "Container stopped."
        fi
    fi
    exit 0
fi

# Local mode — check if lnd is running.
LND_PID=$(pgrep -x lnd 2>/dev/null || true)
if [ -z "$LND_PID" ]; then
    echo "lnd is not running."
    exit 0
fi

echo "Stopping lnd (PID: $LND_PID)..."

if [ "$FORCE" = true ]; then
    kill "$LND_PID"
    echo "Sent SIGTERM."
else
    # Try graceful shutdown via lncli.
    if lncli --network="$NETWORK" --lnddir="$LND_DIR" stop 2>/dev/null; then
        echo "Graceful shutdown initiated."
    else
        echo "lncli stop failed, sending SIGTERM..."
        kill "$LND_PID"
    fi
fi

# Wait for process to exit.
echo "Waiting for lnd to exit..."
for i in {1..15}; do
    if ! kill -0 "$LND_PID" 2>/dev/null; then
        echo "lnd stopped."
        exit 0
    fi
    sleep 1
done

echo "Warning: lnd did not exit within 15 seconds." >&2
echo "Use --force or kill -9 $LND_PID" >&2
exit 1
