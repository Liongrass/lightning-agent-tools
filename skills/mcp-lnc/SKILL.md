---
name: mcp-lnc
description: Build and configure the MCP server for Lightning Node Connect (LNC). Connects AI assistants to lnd nodes via encrypted WebSocket tunnels using pairing phrases — no direct network access or TLS certs needed. Read-only by default (18 tools for querying node state, channels, payments, invoices, peers, on-chain data).
---

# MCP LNC Server

Build and configure the MCP server that connects AI assistants to Lightning
nodes via **Lightning Node Connect (LNC)**. LNC uses encrypted WebSocket tunnels
through a mailbox relay, so the agent never needs direct gRPC access, TLS
certificates, or macaroons — just a 10-word pairing phrase from Lightning
Terminal.

The MCP server is **read-only by default** — it exposes 18 tools for querying
node state but cannot send payments or modify channels.

## Quick Start

```bash
# 1. Build the MCP server binary
skills/mcp-lnc/scripts/install.sh

# 2. Configure environment (mailbox server, dev mode, etc.)
skills/mcp-lnc/scripts/configure.sh

# 3. Add to Claude Code as an MCP server
skills/mcp-lnc/scripts/setup-claude-config.sh
```

Then restart Claude Code. The `lnc_connect` tool will be available to connect
to any lnd node using a pairing phrase.

## How It Works

```
Claude Code  <--stdio-->  mcp-lnc-server  <--LNC WebSocket-->  Mailbox  <-->  lnd
```

1. Claude Code launches `mcp-lnc-server` as a subprocess (stdio transport)
2. Agent calls `lnc_connect` with a pairing phrase and password
3. Server generates an ephemeral ECDSA keypair and opens an encrypted WebSocket
   tunnel through the mailbox relay
4. Once connected, the agent can call any of the 18 read-only tools
5. `lnc_disconnect` closes the tunnel

No keys, certs, or macaroons are stored on disk — the pairing phrase is the
only credential, and it's handled in-memory only.

## Installation

```bash
# Build from source (requires Go 1.24+)
skills/mcp-lnc/scripts/install.sh

# Verify
mcp-lnc-server -version
```

The install script builds from the `mcp-server/` directory in this repo.

## Configuration

```bash
# Generate .env with defaults
skills/mcp-lnc/scripts/configure.sh

# Production (mainnet via Lightning Terminal)
skills/mcp-lnc/scripts/configure.sh --production

# Development (local regtest)
skills/mcp-lnc/scripts/configure.sh --dev --mailbox aperture:11110
```

Configuration is stored in `mcp-server/.env`. Key settings:

| Variable | Default | Description |
|----------|---------|-------------|
| `LNC_MAILBOX_SERVER` | `mailbox.terminal.lightning.today:443` | Mailbox relay server |
| `LNC_DEV_MODE` | `false` | Enable development mode |
| `LNC_INSECURE` | `false` | Skip TLS verification (dev only) |
| `LNC_CONNECT_TIMEOUT` | `30` | Connection timeout in seconds |

## Claude Code Integration

```bash
# Add mcp-lnc-server to Claude Code's MCP config
skills/mcp-lnc/scripts/setup-claude-config.sh

# Project-level config (current project only)
skills/mcp-lnc/scripts/setup-claude-config.sh --scope project

# Global config (all projects)
skills/mcp-lnc/scripts/setup-claude-config.sh --scope global
```

This adds the server to Claude Code's `.mcp.json` (project) or
`~/.claude.json` (global) configuration. After restarting Claude Code, the
LNC tools will be available.

### Manual Configuration

Add to `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "lnc": {
      "command": "mcp-lnc-server",
      "env": {
        "LNC_MAILBOX_SERVER": "mailbox.terminal.lightning.today:443"
      }
    }
  }
}
```

Or run via Docker:

```json
{
  "mcpServers": {
    "lnc": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i", "--network", "host",
        "--env", "LNC_MAILBOX_SERVER",
        "--env", "LNC_DEV_MODE",
        "--env", "LNC_INSECURE",
        "mcp-lnc-server"
      ]
    }
  }
}
```

## Available Tools (18)

### Connection

| Tool | Description |
|------|-------------|
| `lnc_connect` | Connect to lnd via LNC pairing phrase |
| `lnc_disconnect` | Close active LNC connection |

### Node

| Tool | Description |
|------|-------------|
| `lnc_get_info` | Node alias, version, sync status, block height |
| `lnc_get_balance` | Wallet balance (on-chain) and channel balance |

### Channels

| Tool | Description |
|------|-------------|
| `lnc_list_channels` | Active/inactive channels with capacity, balances |
| `lnc_pending_channels` | Channels being opened or closed |

### Invoices

| Tool | Description |
|------|-------------|
| `lnc_decode_invoice` | Decode a BOLT11 invoice |
| `lnc_list_invoices` | List invoices with pagination |
| `lnc_lookup_invoice` | Look up invoice by payment hash |

### Payments

| Tool | Description |
|------|-------------|
| `lnc_list_payments` | Payment history with pagination |
| `lnc_track_payment` | Track specific payment by hash |

### Peers & Network

| Tool | Description |
|------|-------------|
| `lnc_list_peers` | Connected peers with stats |
| `lnc_describe_graph` | Lightning Network topology sample |
| `lnc_get_node_info` | Detailed info about a specific node |

### On-Chain

| Tool | Description |
|------|-------------|
| `lnc_list_unspent` | UTXOs with confirmations |
| `lnc_get_transactions` | On-chain transaction history |
| `lnc_estimate_fee` | Fee estimates for confirmation targets |

## Security Model

- **No stored credentials:** Pairing phrase is handled in-memory only. Ephemeral
  ECDSA keypairs are generated per session.
- **Read-only:** No payment, channel, or state-changing operations are exposed.
  The agent can observe but not modify.
- **Encrypted tunnels:** All traffic is encrypted end-to-end through the mailbox
  relay. The mailbox cannot read the traffic.
- **No direct access:** The agent machine never connects directly to the lnd
  node's gRPC port — all traffic goes through the mailbox.

### Comparison with Direct gRPC Access

| | MCP LNC Server | Direct lncli/gRPC |
|---|---|---|
| **Credential** | Pairing phrase (in-memory) | TLS cert + macaroon (on disk) |
| **Network** | WebSocket via mailbox relay | Direct TCP to gRPC port |
| **Firewall** | No inbound ports needed | Port 10009 must be reachable |
| **Permissions** | Read-only (hardcoded) | Depends on macaroon scope |
| **Setup** | Pairing phrase from Lightning Terminal | Export cert + macaroon files |

## Prerequisites

- **Go 1.24+** for building from source
- **Lightning Terminal (litd)** on the target node for generating pairing phrases
- **Claude Code** for MCP integration

## Troubleshooting

### "pairing phrase must be exactly 10 words"
The pairing phrase is generated by Lightning Terminal. It must be exactly 10
space-separated words.

### "connection timeout"
Check that the mailbox server is reachable. For production, ensure
`mailbox.terminal.lightning.today:443` is not blocked by a firewall.

### "TLS handshake failure"
If using a local regtest setup, enable dev mode and insecure mode:
```bash
skills/mcp-lnc/scripts/configure.sh --dev --insecure
```

### Tools not appearing in Claude Code
Restart Claude Code after running `setup-claude-config.sh`. Check that
`mcp-lnc-server` is on your `$PATH`:
```bash
which mcp-lnc-server
```
