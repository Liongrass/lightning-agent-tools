#!/usr/bin/env bash
# Set up a remote signer: create wallet, export credentials bundle.
#
# Usage:
#   setup-signer.sh                          # Default (mainnet)
#   setup-signer.sh --network testnet        # Testnet
#   setup-signer.sh --password "mypass"      # Custom passphrase
#
# Creates:
#   ~/.lnget/signer/wallet-password.txt      (mode 0600)
#   ~/.lnget/signer/seed.txt                 (mode 0600)
#   ~/.lnget/signer/signer-lnd.conf
#   ~/.lnget/signer/credentials-bundle/      (exported credentials)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LNGET_SIGNER_DIR="${LNGET_SIGNER_DIR:-$HOME/.lnget/signer}"
LND_SIGNER_DIR="${LND_SIGNER_DIR:-$HOME/.lnd-signer}"
NETWORK="mainnet"
PASSWORD=""
RPC_PORT=10012
REST_PORT=10013

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
        --password)
            PASSWORD="$2"
            shift 2
            ;;
        --rpc-port)
            RPC_PORT="$2"
            shift 2
            ;;
        --rest-port)
            REST_PORT="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: setup-signer.sh [options]"
            echo ""
            echo "Set up an lnd remote signer node."
            echo ""
            echo "Options:"
            echo "  --network NETWORK   Bitcoin network (default: mainnet)"
            echo "  --lnddir DIR        Signer lnd data directory (default: ~/.lnd-signer)"
            echo "  --password PASS     Wallet passphrase (auto-generated if omitted)"
            echo "  --rpc-port PORT     Signer RPC port (default: 10012)"
            echo "  --rest-port PORT    Signer REST port (default: 10013)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

echo "=== Remote Signer Setup ==="
echo ""
echo "Network:     $NETWORK"
echo "Signer dir:  $LND_SIGNER_DIR"
echo "Creds dir:   $LNGET_SIGNER_DIR"
echo "RPC port:    $RPC_PORT"
echo "REST port:   $REST_PORT"
echo ""

# Verify lnd is installed.
if ! command -v lnd &>/dev/null; then
    echo "Error: lnd not found. Run install.sh first." >&2
    exit 1
fi

# Create directories with restricted permissions.
mkdir -p "$LNGET_SIGNER_DIR"
chmod 700 "$LNGET_SIGNER_DIR"
mkdir -p "$LND_SIGNER_DIR"

PASSWORD_FILE="$LNGET_SIGNER_DIR/wallet-password.txt"
SEED_OUTPUT="$LNGET_SIGNER_DIR/seed.txt"
CONF_FILE="$LNGET_SIGNER_DIR/signer-lnd.conf"

# Generate or use provided passphrase.
if [ -n "$PASSWORD" ]; then
    echo "Using provided passphrase."
else
    echo "Generating secure passphrase..."
    PASSWORD=$(openssl rand -base64 32 | tr -d '/+=' | head -c 32)
fi

# Store passphrase with restricted permissions.
echo -n "$PASSWORD" > "$PASSWORD_FILE"
chmod 600 "$PASSWORD_FILE"
echo "Passphrase saved to $PASSWORD_FILE (mode 0600)"
echo ""

# Create signer config from template.
TEMPLATE="$SCRIPT_DIR/../templates/signer-lnd.conf.template"
if [ -f "$TEMPLATE" ]; then
    echo "Creating signer config from template..."
    sed "s/bitcoin\.mainnet=true/bitcoin.$NETWORK=true/g" "$TEMPLATE" > "$CONF_FILE"

    # Replace password file path.
    sed -i.bak "s|wallet-unlock-password-file=.*|wallet-unlock-password-file=$PASSWORD_FILE|g" "$CONF_FILE"

    # Replace port numbers.
    sed -i.bak "s|rpclisten=0.0.0.0:10012|rpclisten=0.0.0.0:$RPC_PORT|g" "$CONF_FILE"
    sed -i.bak "s|restlisten=localhost:10013|restlisten=localhost:$REST_PORT|g" "$CONF_FILE"
    rm -f "$CONF_FILE.bak"
    echo "Config saved to $CONF_FILE"
else
    echo "Error: Config template not found at $TEMPLATE" >&2
    exit 1
fi
echo ""

# Start signer lnd temporarily for wallet creation if not running.
SIGNER_WAS_RUNNING=false
if curl -sk "https://localhost:$REST_PORT/v1/state" &>/dev/null; then
    SIGNER_WAS_RUNNING=true
    echo "Signer lnd is already running."
else
    echo "Starting signer lnd temporarily for wallet creation..."
    nohup lnd \
        --lnddir="$LND_SIGNER_DIR" \
        --configfile="$CONF_FILE" \
        > "$LNGET_SIGNER_DIR/signer-setup.log" 2>&1 &
    SIGNER_PID=$!

    echo "Waiting for signer lnd to start (PID: $SIGNER_PID)..."
    for i in {1..30}; do
        if curl -sk "https://localhost:$REST_PORT/v1/state" &>/dev/null; then
            break
        fi
        if ! kill -0 "$SIGNER_PID" 2>/dev/null; then
            echo "Error: signer lnd exited. Check $LNGET_SIGNER_DIR/signer-setup.log" >&2
            exit 1
        fi
        sleep 2
        echo "  Waiting... ($i/30)"
    done
    echo ""
fi

# Create wallet via REST API.
echo "=== Creating Signer Wallet ==="

# Generate seed.
echo "Generating wallet seed..."
SEED_RESPONSE=$(curl -sk -X GET \
    "https://localhost:$REST_PORT/v1/genseed" 2>&1)

MNEMONIC=$(echo "$SEED_RESPONSE" | jq -r '.cipher_seed_mnemonic[]' 2>/dev/null)
if [ -z "$MNEMONIC" ] || [ "$MNEMONIC" = "null" ]; then
    echo "Error: Failed to generate seed." >&2
    echo "Response: $SEED_RESPONSE" >&2
    exit 1
fi

# Store seed with restricted permissions.
echo "$MNEMONIC" > "$SEED_OUTPUT"
chmod 600 "$SEED_OUTPUT"
echo "Seed mnemonic saved to $SEED_OUTPUT (mode 0600)"
echo ""

# Initialize wallet with password and seed.
SEED_JSON=$(echo "$MNEMONIC" | jq -R . | jq -s .)
PAYLOAD=$(jq -n \
    --arg pass "$(echo -n "$PASSWORD" | base64)" \
    --argjson seed "$SEED_JSON" \
    '{wallet_password: $pass, cipher_seed_mnemonic: $seed}')

RESPONSE=$(curl -sk -X POST \
    "https://localhost:$REST_PORT/v1/initwallet" \
    -H "Content-Type: application/json" \
    -d "$PAYLOAD" 2>&1)

ERROR=$(echo "$RESPONSE" | jq -r '.message // empty' 2>/dev/null)
if [ -n "$ERROR" ]; then
    echo "Error creating wallet: $ERROR" >&2
    exit 1
fi

echo "Signer wallet created successfully!"
echo ""

# Wait for signer to be fully ready (wallet unlocked, RPC available).
echo "Waiting for signer to be fully ready..."
for i in {1..30}; do
    STATE=$(curl -sk "https://localhost:$REST_PORT/v1/state" 2>/dev/null | jq -r '.state // empty' 2>/dev/null)
    if [ "$STATE" = "SERVER_ACTIVE" ]; then
        break
    fi
    sleep 2
    echo "  Waiting for RPC... ($i/30)"
done
echo ""

# Export credentials bundle.
echo "=== Exporting Credentials Bundle ==="
"$SCRIPT_DIR/export-credentials.sh" \
    --network "$NETWORK" \
    --lnddir "$LND_SIGNER_DIR" \
    --rpc-port "$RPC_PORT"

echo ""
echo "=== Signer Setup Complete ==="
echo ""
echo "Credential locations:"
echo "  Passphrase: $PASSWORD_FILE"
echo "  Seed:       $SEED_OUTPUT"
echo "  Config:     $CONF_FILE"
echo ""
echo "IMPORTANT: The seed mnemonic at $SEED_OUTPUT is the master secret."
echo "Back it up securely and restrict access to this machine."
echo ""
echo "Next steps:"
echo "  1. Copy the credentials bundle to your agent machine"
echo "  2. On the agent: skills/lnd/scripts/import-credentials.sh --bundle <path>"
echo "  3. On the agent: skills/lnd/scripts/create-wallet.sh"
echo "  4. On the agent: skills/lnd/scripts/start-lnd.sh --signer-host <this-ip>:$RPC_PORT"
