#!/usr/bin/env bash
# Wrapper for lncli with auto-detected paths.
#
# Usage:
#   lncli.sh getinfo
#   lncli.sh walletbalance
#   lncli.sh --network testnet getinfo
#   lncli.sh --container sam getinfo               # Run inside Docker container
#   lncli.sh openchannel --node_key=<pubkey> --local_amt=1000000

set -e

LND_DIR="${LND_DIR:-}"
NETWORK="${NETWORK:-mainnet}"
CONTAINER=""
LNCLI_ARGS=()

# Parse our arguments (pass everything else to lncli).
while [[ $# -gt 0 ]]; do
    case $1 in
        --network)
            NETWORK="$2"
            shift 2
            ;;
        --lnddir)
            LND_DIR="$2"
            shift 2
            ;;
        --container)
            CONTAINER="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: lncli.sh [--network NET] [--lnddir DIR] [--container NAME] <command> [args]"
            echo ""
            echo "Wrapper for lncli with auto-detected paths."
            echo ""
            echo "Options:"
            echo "  --network NETWORK    Bitcoin network (default: mainnet)"
            echo "  --lnddir DIR         lnd data directory (default: ~/.lnd)"
            echo "  --container NAME     Run lncli inside a Docker container"
            echo ""
            echo "All other arguments are passed directly to lncli."
            exit 0
            ;;
        *)
            LNCLI_ARGS+=("$1")
            shift
            ;;
    esac
done

if [ ${#LNCLI_ARGS[@]} -eq 0 ]; then
    echo "Error: No lncli command specified." >&2
    echo "Usage: lncli.sh <command> [args]" >&2
    exit 1
fi

# Apply default lnddir if not set.
if [ -z "$LND_DIR" ]; then
    if [ -n "$CONTAINER" ]; then
        LND_DIR="/root/.lnd"
    else
        LND_DIR="$HOME/.lnd"
    fi
fi

# Verify lncli is available.
if [ -n "$CONTAINER" ]; then
    if ! docker exec "$CONTAINER" which lncli &>/dev/null; then
        echo "Error: lncli not found in container '$CONTAINER'." >&2
        exit 1
    fi
    exec docker exec "$CONTAINER" lncli --network="$NETWORK" --lnddir="$LND_DIR" "${LNCLI_ARGS[@]}"
else
    if ! command -v lncli &>/dev/null; then
        echo "Error: lncli not found. Run install.sh first." >&2
        exit 1
    fi
    exec lncli --network="$NETWORK" --lnddir="$LND_DIR" "${LNCLI_ARGS[@]}"
fi
