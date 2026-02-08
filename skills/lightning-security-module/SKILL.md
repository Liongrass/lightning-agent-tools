---
name: lightning-security-module
description: Set up an lnd remote signer node that holds private keys on a separate secure machine. Exports a credentials bundle (accounts JSON, TLS cert, admin macaroon) for watch-only nodes to consume. Use when firewalling private key material from AI agents.
---

# Lightning Security Module (Remote Signer)

Set up an lnd remote signer node that holds private keys on a separate, secured
machine. The signer never routes payments or opens channels — it only holds
keys and signs when asked by a watch-only lnd node.

## Architecture

```
Agent Machine                     Signer Machine (secure)
┌─────────────────┐              ┌─────────────────────┐
│  lnd (watch-only)│◄──gRPC────►│  lnd (signer)        │
│  - neutrino      │             │  - holds seed         │
│  - manages chans │             │  - signs commitments  │
│  - routes pmts   │             │  - signs on-chain txs │
│  - NO key material│            │  - no p2p networking   │
└─────────────────┘              └─────────────────────┘
```

The watch-only node handles all networking and channel management. The signer
node holds the seed and performs cryptographic signing. Even if the agent machine
is fully compromised, the attacker cannot extract private keys or sign arbitrary
transactions.

See [references/architecture.md](references/architecture.md) for the full
architecture explainer.

## Quick Start

Run these on the **signer machine** (the secure machine that holds keys):

```bash
# 1. Install lnd on the signer
skills/lightning-security-module/scripts/install.sh

# 2. Set up signer wallet and export credentials bundle
skills/lightning-security-module/scripts/setup-signer.sh

# 3. Copy the credentials bundle to your agent machine
#    The setup script prints the bundle path and a base64 string for easy transfer.
```

Then on the **agent machine**, use the lnd skill to import and run watch-only:

```bash
# 4. Import credentials bundle
skills/lnd/scripts/import-credentials.sh --bundle <credentials-bundle>

# 5. Create watch-only wallet (connects to signer during creation)
skills/lnd/scripts/create-wallet.sh --signer-host <signer-ip>:10012

# 6. Start lnd in watch-only mode (connects to remote signer)
skills/lnd/scripts/start-lnd.sh --signer-host <signer-ip>:10012
```

## Credential Bundle Format

The exported bundle (`~/.lnget/signer/credentials-bundle/`) contains:

| File | Purpose |
|------|---------|
| `accounts.json` | Account xpubs for watch-only wallet import |
| `tls.cert` | Signer's TLS certificate for authenticated gRPC |
| `admin.macaroon` | Signer's admin macaroon for RPC authentication |

The bundle is also available as a single base64-encoded tar.gz file
(`credentials-bundle.tar.gz.b64`) for easy copy-paste transfer between machines.

## Scripts

| Script | Purpose |
|--------|---------|
| `install.sh` | Install lnd on signer machine |
| `setup-signer.sh` | Create signer wallet and export credentials |
| `start-signer.sh` | Start signer lnd |
| `stop-signer.sh` | Stop signer lnd |
| `export-credentials.sh` | Re-export credentials from running signer |

## Managing the Signer

### Start the signer

```bash
skills/lightning-security-module/scripts/start-signer.sh
```

### Stop the signer

```bash
skills/lightning-security-module/scripts/stop-signer.sh
```

### Re-export credentials

If TLS certificates or macaroons have been regenerated:

```bash
skills/lightning-security-module/scripts/export-credentials.sh
```

## Configuration

The signer config template is at `templates/signer-lnd.conf.template`. Key
differences from a standard lnd node:

- **No p2p listening** (`listen=` empty) — signer doesn't route
- **RPC on 0.0.0.0:10012** — accepts connections from watch-only node
- **REST on localhost:10013** — local only, for wallet creation
- **TLS extra IP 0.0.0.0** — so watch-only on a different machine can connect
- **No autopilot, no routing fees** — signer is signing-only

## Security Model

**What stays on the signer:**
- 24-word seed mnemonic
- All private keys (funding, revocation, HTLC)
- Wallet database with key material

**What gets exported:**
- Account xpubs (public keys only — cannot spend)
- TLS certificate (for authenticated connection)
- Admin macaroon (for RPC auth — scope this down for production)

**Threat model:**
- Compromised agent machine cannot sign transactions or extract keys
- Attacker with agent access can see balances and channel state but not spend
- Signer machine should have minimal attack surface (no unnecessary services)

**Production hardening:**
- Replace admin macaroon with a custom least-privilege macaroon
- Restrict signer RPC to specific IP addresses via firewall
- Run signer on dedicated hardware or a hardened VM
- Enable macaroon rotation on a schedule

## Macaroon Bakery for Signer

By default, the credentials bundle includes the signer's `admin.macaroon`. For
production, bake a scoped macaroon that only allows signing operations:

```bash
# On the signer machine — bake a signing-only macaroon
skills/lightning-security-module/scripts/lncli-signer.sh bakemacaroon \
    uri:/signrpc.Signer/SignOutputRaw \
    uri:/signrpc.Signer/ComputeInputScript \
    uri:/signrpc.Signer/MuSig2Sign \
    uri:/walletrpc.WalletKit/DeriveKey \
    uri:/walletrpc.WalletKit/DeriveNextKey \
    --save_to=~/.lnd-signer/data/chain/bitcoin/mainnet/signer-only.macaroon
```

Or use `lncli` directly:

```bash
lncli --rpcserver=localhost:10012 --lnddir=~/.lnd-signer \
    bakemacaroon \
    uri:/signrpc.Signer/SignOutputRaw \
    uri:/signrpc.Signer/ComputeInputScript \
    uri:/signrpc.Signer/MuSig2Sign \
    uri:/walletrpc.WalletKit/DeriveKey \
    uri:/walletrpc.WalletKit/DeriveNextKey \
    --save_to=~/.lnd-signer/data/chain/bitcoin/mainnet/signer-only.macaroon
```

Then re-export the credentials bundle, replacing `admin.macaroon` with the
scoped macaroon in `~/.lnget/signer/credentials-bundle/`.

**Inspect a macaroon's permissions:**

```bash
lncli --rpcserver=localhost:10012 --lnddir=~/.lnd-signer \
    printmacaroon --macaroon_file <path>
```

**List all available permissions for baking:**

```bash
lncli --rpcserver=localhost:10012 --lnddir=~/.lnd-signer listpermissions
```

## Ports

| Port  | Service | Interface   | Description |
|-------|---------|-------------|-------------|
| 10012 | gRPC    | 0.0.0.0     | Signer RPC (watch-only connects here) |
| 10013 | REST    | localhost   | Local REST for wallet creation |

## File Locations

| Path | Purpose |
|------|---------|
| `~/.lnget/signer/signer-lnd.conf` | Signer config |
| `~/.lnget/signer/wallet-password.txt` | Signer wallet passphrase (0600) |
| `~/.lnget/signer/seed.txt` | Signer seed mnemonic (0600) |
| `~/.lnget/signer/credentials-bundle/` | Exported credentials |
| `~/.lnd-signer/` | Signer lnd data directory |
