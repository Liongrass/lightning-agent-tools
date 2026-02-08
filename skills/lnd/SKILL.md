---
name: lnd
description: Install and run lnd Lightning Network daemon natively with neutrino backend and SQLite storage. Defaults to watch-only mode with remote signer for secure agent operation. Use when setting up a Lightning node for payments, managing wallets, opening channels, paying invoices, or enabling an agent to send/receive Lightning payments for L402 commerce.
---

# LND Lightning Network Node

Install and operate an lnd Lightning Network node for agent-driven payments.
Defaults to neutrino (light client) backend with SQLite storage for minimal
setup — no full Bitcoin node required.

**Default mode: watch-only with remote signer.** Private keys stay on a
separate signer machine — the agent never touches key material. For quick
testing, use `--mode standalone` (keys on disk, less secure).

## Quick Start (Watch-Only — Recommended)

Requires a signer set up with the `lightning-security-module` skill.

```bash
# 1. Install lnd
skills/lnd/scripts/install.sh

# 2. Import credentials from signer (see lightning-security-module skill)
skills/lnd/scripts/import-credentials.sh --bundle <credentials-bundle>

# 3. Create watch-only wallet (needs signer host — lnd must connect to signer during wallet creation)
skills/lnd/scripts/create-wallet.sh --signer-host <signer-ip>:10012

# 4. Start lnd (connects to remote signer)
skills/lnd/scripts/start-lnd.sh --signer-host <signer-ip>:10012

# 5. Check status
skills/lnd/scripts/lncli.sh getinfo
```

## Quick Start (Standalone — Testing Only)

For quick testing where security is not a concern. Keys are stored on disk.

```bash
# 1. Install lnd
skills/lnd/scripts/install.sh

# 2. Create wallet (generates seed locally)
skills/lnd/scripts/create-wallet.sh --mode standalone

# 3. Start lnd
skills/lnd/scripts/start-lnd.sh --mode standalone

# 4. Check status
skills/lnd/scripts/lncli.sh getinfo
```

> **Warning:** Standalone mode stores the seed mnemonic and wallet passphrase on
> disk. Any process running as the same user can read them. Do not use for
> mainnet funds you cannot afford to lose.

## Docker

If lnd is running in a Docker container, most scripts accept `--container`:

```bash
# Run lncli commands against a container
skills/lnd/scripts/lncli.sh --container sam getinfo
skills/lnd/scripts/lncli.sh --container sam walletbalance

# Stop lnd in a container
skills/lnd/scripts/stop-lnd.sh --container sam
```

## Remote Nodes

To connect to a remote lnd node, provide the connection credentials:

```bash
# Run lncli against a remote node
skills/lnd/scripts/lncli.sh \
    --rpcserver remote-host:10009 \
    --tlscertpath ~/remote-tls.cert \
    --macaroonpath ~/remote-admin.macaroon \
    getinfo

# Stop a remote node
skills/lnd/scripts/stop-lnd.sh \
    --rpcserver remote-host:10009 \
    --tlscertpath ~/remote-tls.cert \
    --macaroonpath ~/remote-admin.macaroon
```

You need lncli installed locally and copies of the node's TLS cert and macaroon.

## Installation

The install script builds lnd from source with all required build tags:

```bash
skills/lnd/scripts/install.sh
```

This will:
- Verify Go is installed (required)
- Run `go install` with tags: `signrpc walletrpc chainrpc invoicesrpc routerrpc
  peersrpc kvdb_sqlite neutrinorpc`
- Verify `lnd` and `lncli` are on `$PATH`

To install manually:

```bash
go install -tags "signrpc walletrpc chainrpc invoicesrpc routerrpc peersrpc kvdb_sqlite neutrinorpc" github.com/lightningnetwork/lnd/cmd/lnd@latest
go install -tags "signrpc walletrpc chainrpc invoicesrpc routerrpc peersrpc kvdb_sqlite neutrinorpc" github.com/lightningnetwork/lnd/cmd/lncli@latest
```

## Wallet Setup

### Watch-Only Wallet (Default)

Imports account xpubs from the remote signer — no seed or private keys on this
machine.

```bash
# Import credentials bundle from signer
skills/lnd/scripts/import-credentials.sh --bundle <credentials-bundle>

# Create watch-only wallet
skills/lnd/scripts/create-wallet.sh
```

The credentials bundle is produced by the `lightning-security-module` skill's
`export-credentials.sh` script. It contains:
- `accounts.json` — account xpubs for watch-only import
- `tls.cert` — signer's TLS certificate
- `admin.macaroon` — signer's admin macaroon

### Standalone Wallet

Generates a seed locally and stores it on disk. Use only for testing.

```bash
skills/lnd/scripts/create-wallet.sh --mode standalone
```

This handles the full wallet creation flow:

1. Generates a secure random wallet passphrase
2. Starts lnd temporarily (if not running)
3. Calls `lncli create` with the passphrase
4. Captures and stores the 24-word seed mnemonic
5. Stores credentials securely:
   - `~/.lnget/lnd/wallet-password.txt` (mode 0600) — wallet unlock passphrase
   - `~/.lnget/lnd/seed.txt` (mode 0600) — 24-word recovery mnemonic

**Options:**

```bash
# Custom data directory
create-wallet.sh --lnddir ~/.lnd-agent

# Specific network
create-wallet.sh --network mainnet

# Custom passphrase (instead of auto-generated)
create-wallet.sh --mode standalone --password "your-passphrase-here"
```

### Unlock Wallet

After lnd restarts, the wallet must be unlocked before the node is operational:

```bash
skills/lnd/scripts/unlock-wallet.sh
```

This reads the passphrase from `~/.lnget/lnd/wallet-password.txt` and calls the
lnd REST API to unlock. Alternatively, lnd can auto-unlock on start using the
`wallet-unlock-password-file` config option (included in the default template).

### Recover Wallet from Seed (Standalone Only)

```bash
skills/lnd/scripts/create-wallet.sh --mode standalone --recover --seed-file ~/.lnget/lnd/seed.txt
```

## Starting and Stopping

### Start lnd

```bash
# Watch-only (default) — requires signer host
skills/lnd/scripts/start-lnd.sh --signer-host <signer-ip>:10012

# Standalone mode
skills/lnd/scripts/start-lnd.sh --mode standalone
```

Starts lnd as a background process using the config at `~/.lnget/lnd/lnd.conf`.
Defaults:
- **Backend:** neutrino (BIP 157/158 light client)
- **Database:** SQLite
- **Network:** mainnet (override with `--network testnet`)
- **Auto-unlock:** enabled via password file

**Options:**

```bash
# Specify network
start-lnd.sh --network testnet

# Custom lnd directory
start-lnd.sh --lnddir ~/.lnd-agent

# Foreground mode (for debugging)
start-lnd.sh --foreground

# With extra lnd flags
start-lnd.sh --extra-args "--debuglevel=trace"

# Set signer host via environment variable
LND_SIGNER_HOST=10.0.0.5:10012 start-lnd.sh
```

### Stop lnd

```bash
skills/lnd/scripts/stop-lnd.sh
```

Gracefully stops lnd via `lncli stop`. Falls back to SIGTERM if lncli fails.

## Node Operations

All commands go through the lncli wrapper which auto-detects paths and network:

### Node Info

```bash
# Get node status
skills/lnd/scripts/lncli.sh getinfo

# Wallet balance (on-chain)
skills/lnd/scripts/lncli.sh walletbalance

# Channel balance (Lightning)
skills/lnd/scripts/lncli.sh channelbalance
```

### Funding the Wallet

```bash
# Generate a new address
skills/lnd/scripts/lncli.sh newaddress p2tr

# Check balance after sending funds
skills/lnd/scripts/lncli.sh walletbalance
```

For testnet, use a faucet. For mainnet, send BTC to the generated address.

### Channel Management

```bash
# Connect to a peer
skills/lnd/scripts/lncli.sh connect <pubkey>@<host>:9735

# Open a channel (satoshis)
skills/lnd/scripts/lncli.sh openchannel --node_key=<pubkey> --local_amt=1000000

# List channels
skills/lnd/scripts/lncli.sh listchannels

# Check channel balance
skills/lnd/scripts/lncli.sh channelbalance

# Close channel cooperatively
skills/lnd/scripts/lncli.sh closechannel --funding_txid=<txid> --output_index=<n>
```

### Payments

```bash
# Create an invoice
skills/lnd/scripts/lncli.sh addinvoice --amt=1000 --memo="test payment"

# Decode a BOLT11 invoice
skills/lnd/scripts/lncli.sh decodepayreq <bolt11_invoice>

# Pay an invoice
skills/lnd/scripts/lncli.sh sendpayment --pay_req=<bolt11_invoice>

# List payments
skills/lnd/scripts/lncli.sh listpayments

# List received invoices
skills/lnd/scripts/lncli.sh listinvoices
```

### Macaroon Bakery

lnd uses macaroons for API authentication. **Never give agents the admin
macaroon in production.** Use the `macaroon-bakery` skill to bake
least-privilege macaroons for each agent role:

```bash
# Bake a pay-only macaroon
skills/macaroon-bakery/scripts/bake.sh --role pay-only

# Bake an invoice-only macaroon
skills/macaroon-bakery/scripts/bake.sh --role invoice-only

# Inspect any macaroon
skills/macaroon-bakery/scripts/bake.sh --inspect <path-to-macaroon>
```

See the `macaroon-bakery` skill for preset roles, custom permissions, rotation,
and best practices.

**Built-in macaroons** (auto-generated by lnd):

| Macaroon | Capabilities |
|----------|-------------|
| `admin.macaroon` | Full access (read, write, generate invoices, send payments) |
| `readonly.macaroon` | Read-only access (getinfo, balances, list operations) |
| `invoice.macaroon` | Create and manage invoices only |

### Peer Management

```bash
# List connected peers
skills/lnd/scripts/lncli.sh listpeers

# Disconnect from peer
skills/lnd/scripts/lncli.sh disconnect <pubkey>
```

## Configuration

The default config template lives at `skills/lnd/templates/lnd.conf.template`.
On first run, `start-lnd.sh` copies it to `~/.lnget/lnd/lnd.conf`.

Key defaults:

```ini
[Application Options]
alias=lnget-agent
listen=0.0.0.0:9735
rpclisten=localhost:10009
restlisten=localhost:8080
wallet-unlock-password-file=~/.lnget/lnd/wallet-password.txt
wallet-unlock-allow-create=true

[Bitcoin]
bitcoin.active=true
bitcoin.mainnet=true
bitcoin.node=neutrino

[neutrino]
neutrino.addpeer=btcd0.lightning.computer
neutrino.addpeer=mainnet1-btcd.zaphq.io
neutrino.addpeer=mainnet2-btcd.zaphq.io
neutrino.feeurl=https://nodes.lightning.computer/fees/v1/btc-fee-estimates.json

[db]
db.backend=sqlite
```

In watch-only mode, `start-lnd.sh` automatically appends:

```ini
[remotesigner]
remotesigner.enable=true
remotesigner.rpchost=<signer-host>
remotesigner.tlscertpath=~/.lnget/lnd/signer-credentials/tls.cert
remotesigner.macaroonpath=~/.lnget/lnd/signer-credentials/admin.macaroon
```

Override network:

```bash
# For testnet
start-lnd.sh --network testnet
```

## Ports

| Port  | Service   | Description                    |
|-------|-----------|--------------------------------|
| 9735  | Lightning | Peer-to-peer Lightning Network |
| 10009 | gRPC      | lncli and programmatic access  |
| 8080  | REST      | REST API (wallet unlock, etc.) |

## File Locations

| Path | Purpose |
|------|---------|
| `~/.lnget/lnd/lnd.conf` | Configuration file |
| `~/.lnget/lnd/wallet-password.txt` | Wallet unlock passphrase (0600) |
| `~/.lnget/lnd/seed.txt` | 24-word mnemonic backup (0600, standalone only) |
| `~/.lnget/lnd/signer-credentials/` | Imported signer credentials (watch-only) |
| `~/.lnd/` | lnd data directory (default) |
| `~/.lnd/data/chain/bitcoin/<network>/` | Chain data and macaroons |
| `~/.lnd/tls.cert` | TLS certificate |
| `~/.lnd/tls.key` | TLS private key |
| `~/.lnd/logs/` | Log files |

## Integration with lnget

Once lnd is running with a funded wallet and open channels, configure lnget to
use it:

```bash
# Initialize lnget config
lnget config init

# lnget auto-detects lnd at localhost:10009 with default paths
lnget ln status

# Fetch an L402-protected resource
lnget --max-cost 1000 https://api.example.com/paid-data
```

Or set config explicitly:

```yaml
# ~/.lnget/config.yaml
ln:
  mode: lnd
  lnd:
    host: localhost:10009
    tls_cert: ~/.lnd/tls.cert
    macaroon: ~/.lnd/data/chain/bitcoin/mainnet/admin.macaroon
    network: mainnet
```

**For agents using baked macaroons**, point to the custom macaroon instead:

```yaml
ln:
  mode: lnd
  lnd:
    host: localhost:10009
    tls_cert: ~/.lnd/tls.cert
    macaroon: ~/.lnd/data/chain/bitcoin/mainnet/pay-only.macaroon
    network: mainnet
```

## Security Considerations

See [references/security.md](references/security.md) for detailed security
guidance.

**Default model (watch-only with remote signer):**
- No seed or private keys on the agent machine
- Signing delegated to a separate signer node via gRPC
- Credentials bundle (xpubs, TLS cert, macaroon) imported from signer
- Set up with the `lightning-security-module` skill

**Standalone model (testing only):**
- Wallet passphrase stored on disk at `~/.lnget/lnd/wallet-password.txt`
- Seed mnemonic stored on disk at `~/.lnget/lnd/seed.txt`
- Both files created with mode 0600 (owner read/write only)
- Suitable for testnet, small amounts, and quick testing

**Macaroon security:**
- Never give agents the admin macaroon in production
- Bake custom macaroons with minimum required permissions
- Use `bakemacaroon` to create scoped credentials for each agent role
- See the Macaroon Bakery section above for examples

## Troubleshooting

### "wallet not found"
Run `skills/lnd/scripts/create-wallet.sh` to create the wallet first.

### "wallet locked"
Run `skills/lnd/scripts/unlock-wallet.sh` or restart lnd (auto-unlock is
enabled by default in the config template).

### "chain backend is still syncing"
Neutrino needs time to sync headers. Check progress with:
```bash
skills/lnd/scripts/lncli.sh getinfo | jq '{synced_to_chain, block_height}'
```

### "unable to find a path to destination"
No route exists. Check channel balances:
```bash
skills/lnd/scripts/lncli.sh listchannels | jq '.[].channels[] | {remote_pubkey, local_balance, remote_balance}'
```

### "connect: connection refused" on lncli
lnd is not running or not listening. Check:
```bash
skills/lnd/scripts/lncli.sh --help  # Verify lncli works
pgrep lnd                           # Check if lnd process exists
```

### "remote signer not reachable"
The watch-only node cannot connect to the signer. Check:
```bash
# Verify signer is running
curl -sk https://<signer-ip>:10012/v1/state

# Check signer credentials are imported
ls -la ~/.lnget/lnd/signer-credentials/

# Verify TLS cert matches
openssl x509 -in ~/.lnget/lnd/signer-credentials/tls.cert -noout -subject
```
