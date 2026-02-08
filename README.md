# Lightning Agent Kit

A Claude Code plugin that gives AI agents the skills to operate on the Lightning
Network — run nodes, send/receive payments, bake scoped credentials, and host
paid API endpoints.

## Skills

| Skill | Description |
|-------|-------------|
| **lnd** | Install and run an lnd Lightning node. Defaults to watch-only mode with remote signer. |
| **lightning-security-module** | Set up a remote signer node that holds private keys on a separate machine. |
| **macaroon-bakery** | Bake least-privilege macaroons for scoped agent access. |
| **lnget** | Lightning-native HTTP client with automatic L402 payment support. |
| **aperture** | L402 reverse proxy for hosting paid API endpoints. |
| **mcp-lnc** | MCP server for Lightning Node Connect — connects AI assistants to lnd via encrypted WebSocket tunnels. |
| **commerce** | End-to-end agent commerce workflow (lnd + lnget + aperture). |

## Quick Start

### Install the plugin

Clone this repo and point Claude Code at it:

```bash
git clone https://github.com/lightninglabs/lightning-agent-kit.git
cd lightning-agent-kit
```

Claude Code discovers skills via `.claude/skills/` automatically.

### Example prompts

```
Get node info for sam in Docker regtest
```

```
Bake a pay-only macaroon on zane in Docker regtest
```

```
Export credentials from my signer and bake a signer-only macaroon
```

```
Connect to my remote node at host:10009 and check the wallet balance
```

### Docker containers

All scripts support `--container` for lnd nodes running in Docker:

```bash
skills/lnd/scripts/lncli.sh --container sam --network regtest getinfo
skills/macaroon-bakery/scripts/bake.sh --role pay-only --container sam --network regtest
```

### Remote nodes

All scripts support `--rpcserver`, `--tlscertpath`, and `--macaroonpath` for
remote lnd nodes (e.g., Voltage):

```bash
skills/lnd/scripts/lncli.sh \
    --rpcserver your-node.voltageapp.io:10009 \
    --tlscertpath ~/tls.cert \
    --macaroonpath ~/admin.macaroon \
    --network mainnet getinfo
```

## Architecture

```
lightning-agent-kit/
├── .claude/skills/          # Symlinks for Claude Code skill discovery
├── .claude-plugin/          # Plugin metadata
├── mcp-server/              # MCP server for Lightning Node Connect
└── skills/
    ├── lnd/                 # Lightning node operations
    ├── lightning-security-module/  # Remote signer setup
    ├── macaroon-bakery/     # Credential scoping
    ├── mcp-lnc/             # MCP server build & config
    ├── lnget/               # L402 HTTP client
    ├── aperture/            # L402 reverse proxy
    └── commerce/            # Full commerce workflow
```

## Security Model

The default setup uses lnd's **remote signer architecture**:

- **Signer machine** holds private keys, never routes payments
- **Agent machine** runs a watch-only lnd node, delegates signing
- Even if the agent machine is compromised, keys cannot be extracted

For credentials, use the **macaroon-bakery** skill to bake least-privilege
macaroons — never give agents `admin.macaroon` in production.

## MCP Server

The `mcp-server/` directory contains an MCP server that connects to lnd via
Lightning Node Connect (LNC). This enables AI assistants to interact with
Lightning nodes through the Model Context Protocol.

Use the `mcp-lnc` skill to build, configure, and wire it into Claude Code:

```bash
skills/mcp-lnc/scripts/install.sh
skills/mcp-lnc/scripts/configure.sh
skills/mcp-lnc/scripts/setup-claude-config.sh
```

## Prerequisites

- **Go 1.21+** for building lnd/lncli from source
- **Docker** (optional) for container-based lnd nodes
- **jq** for JSON processing in scripts

## License

See [LICENSE](mcp-server/LICENSE).
