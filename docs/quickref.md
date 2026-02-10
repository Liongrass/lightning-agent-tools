# Quick Reference

> Every important command in one place.

## Installation

```bash
skills/lnd/scripts/install.sh                              # lnd + lncli
skills/lnget/scripts/install.sh                            # lnget CLI
skills/aperture/scripts/install.sh                         # aperture proxy
skills/mcp-lnc/scripts/install.sh                          # MCP server
skills/lightning-security-module/scripts/install.sh         # lnd (signer machine)
```

## Node Operations

```bash
skills/lnd/scripts/start-lnd.sh                            # start lnd
skills/lnd/scripts/stop-lnd.sh                             # stop lnd
skills/lnd/scripts/lncli.sh getinfo                        # node status
skills/lnd/scripts/lncli.sh walletbalance                  # on-chain balance
skills/lnd/scripts/lncli.sh channelbalance                 # channel balance
skills/lnd/scripts/unlock-wallet.sh                        # unlock after restart
```

## Wallet

```bash
# Watch-only (production, imports xpubs from signer)
skills/lnd/scripts/import-credentials.sh --bundle <path>
skills/lnd/scripts/create-wallet.sh --signer-host <ip>:10012

# Standalone (testing, generates local seed)
skills/lnd/scripts/create-wallet.sh --mode standalone

# Funding
skills/lnd/scripts/lncli.sh newaddress p2tr               # generate address
skills/lnd/scripts/lncli.sh walletbalance                  # check balance
```

## Channels

```bash
skills/lnd/scripts/lncli.sh connect <pubkey>@<host>:9735                      # connect to peer
skills/lnd/scripts/lncli.sh openchannel --node_key=<pubkey> --local_amt=N      # open channel
skills/lnd/scripts/lncli.sh listchannels                                       # list channels
skills/lnd/scripts/lncli.sh pendingchannels                                    # pending opens/closes
skills/lnd/scripts/lncli.sh closechannel --funding_txid=<txid> --output_index=N  # close channel
skills/lnd/scripts/lncli.sh listpeers                                          # connected peers
skills/lnd/scripts/lncli.sh disconnect <pubkey>                                # disconnect peer
```

## Payments

```bash
skills/lnd/scripts/lncli.sh addinvoice --amt=1000 --memo="description"    # create invoice
skills/lnd/scripts/lncli.sh decodepayreq <bolt11>                          # decode invoice
skills/lnd/scripts/lncli.sh sendpayment --pay_req=<bolt11>                 # pay invoice
skills/lnd/scripts/lncli.sh listpayments                                   # payment history
skills/lnd/scripts/lncli.sh listinvoices                                   # invoice history
```

## Macaroon Bakery

```bash
# Preset roles
skills/macaroon-bakery/scripts/bake.sh --role pay-only
skills/macaroon-bakery/scripts/bake.sh --role invoice-only
skills/macaroon-bakery/scripts/bake.sh --role read-only
skills/macaroon-bakery/scripts/bake.sh --role channel-admin
skills/macaroon-bakery/scripts/bake.sh --role signer-only

# Custom
skills/macaroon-bakery/scripts/bake.sh --custom \
    uri:/lnrpc.Lightning/SendPaymentSync \
    uri:/lnrpc.Lightning/DecodePayReq \
    uri:/lnrpc.Lightning/GetInfo

# Inspect
skills/macaroon-bakery/scripts/bake.sh --inspect <path-to-macaroon>

# List all available permissions
skills/macaroon-bakery/scripts/bake.sh --list-permissions

# Save to specific path
skills/macaroon-bakery/scripts/bake.sh --role pay-only --save-to ~/agent.macaroon
```

## lnget

```bash
# Fetch
lnget https://api.example.com/data.json                   # fetch to stdout
lnget -o data.json https://api.example.com/data.json       # fetch to file
lnget -q https://api.example.com/data.json | jq .          # quiet mode, pipe
lnget -X POST -d '{"q":"test"}' https://api.example.com    # POST with body

# Cost control
lnget --max-cost 500 https://api.example.com/data          # max auto-pay amount
lnget --no-pay https://api.example.com/data                # preview without paying
lnget --no-pay --json https://... | jq '.invoice_amount_sat'  # check price

# Tokens
lnget tokens list                                          # list cached tokens
lnget tokens show api.example.com                          # show specific token
lnget tokens remove api.example.com                        # force re-payment
lnget tokens clear --force                                 # clear all tokens

# Configuration
lnget config init                                          # initialize config
lnget config show                                          # show current config

# Backend status
lnget ln status                                            # connection status
lnget ln info                                              # backend info

# LNC pairing
lnget ln lnc pair "ten word pairing phrase here"           # pair with LNC
lnget ln lnc sessions                                      # list LNC sessions
lnget ln lnc revoke <session-id>                           # revoke session

# Neutrino (embedded wallet)
lnget ln neutrino init                                     # initialize
lnget ln neutrino fund                                     # funding address
lnget ln neutrino balance                                  # check balance
```

## Aperture

```bash
skills/aperture/scripts/setup.sh                           # generate config
skills/aperture/scripts/setup.sh --insecure --port 8081    # dev mode
skills/aperture/scripts/setup.sh --network testnet         # testnet
skills/aperture/scripts/start.sh                           # start proxy
skills/aperture/scripts/stop.sh                            # stop proxy
```

## MCP Server

```bash
skills/mcp-lnc/scripts/install.sh                         # build from source
skills/mcp-lnc/scripts/configure.sh                        # generate .env
skills/mcp-lnc/scripts/configure.sh --production           # mainnet config
skills/mcp-lnc/scripts/configure.sh --dev --insecure       # regtest config
skills/mcp-lnc/scripts/setup-claude-config.sh --scope project   # add to .mcp.json
skills/mcp-lnc/scripts/setup-claude-config.sh --scope global    # add to ~/.claude.json
```

## Remote Signer

```bash
# On signer machine
skills/lightning-security-module/scripts/install.sh        # install lnd
skills/lightning-security-module/scripts/setup-signer.sh   # create wallet + export creds
skills/lightning-security-module/scripts/start-signer.sh   # start signer
skills/lightning-security-module/scripts/stop-signer.sh    # stop signer
skills/lightning-security-module/scripts/export-credentials.sh  # re-export bundle

# On agent machine
skills/lnd/scripts/import-credentials.sh --bundle <path>
skills/lnd/scripts/create-wallet.sh --signer-host <ip>:10012
skills/lnd/scripts/start-lnd.sh --signer-host <ip>:10012

# Scope signer macaroon
skills/macaroon-bakery/scripts/bake.sh --role signer-only \
    --rpc-port 10012 --lnddir ~/.lnd-signer
```

## Docker Containers

All `lncli` and bakery commands support `--container` for Docker-based nodes:

```bash
skills/lnd/scripts/lncli.sh --container sam --network regtest getinfo
skills/lnd/scripts/lncli.sh --container sam --network regtest walletbalance
skills/macaroon-bakery/scripts/bake.sh --role pay-only --container sam --network regtest
skills/macaroon-bakery/scripts/bake.sh --inspect /root/.lnd/data/chain/bitcoin/regtest/admin.macaroon --container sam
skills/lnd/scripts/stop-lnd.sh --container sam
skills/lightning-security-module/scripts/export-credentials.sh --container sam --network regtest
```

## Remote Nodes

All scripts support direct connection to remote lnd nodes:

```bash
skills/lnd/scripts/lncli.sh \
    --rpcserver remote-host:10009 \
    --tlscertpath ~/remote-tls.cert \
    --macaroonpath ~/remote-admin.macaroon \
    getinfo

skills/macaroon-bakery/scripts/bake.sh --role pay-only \
    --rpcserver remote-host:10009 \
    --tlscertpath ~/remote-tls.cert \
    --macaroonpath ~/remote-admin.macaroon \
    --save-to ~/remote-pay-only.macaroon
```

## File Paths

| Path | Purpose |
|------|---------|
| `~/.lnget/lnd/lnd.conf` | lnd configuration |
| `~/.lnget/lnd/wallet-password.txt` | Wallet passphrase (0600) |
| `~/.lnget/lnd/seed.txt` | Wallet seed, standalone only (0600) |
| `~/.lnget/lnd/signer-credentials/` | Imported signer credentials |
| `~/.lnget/signer/signer-lnd.conf` | Signer configuration |
| `~/.lnget/signer/wallet-password.txt` | Signer passphrase (0600) |
| `~/.lnget/signer/seed.txt` | Signer seed (0600) |
| `~/.lnget/signer/credentials-bundle/` | Exported signer credentials |
| `~/.lnget/config.yaml` | lnget configuration |
| `~/.lnget/tokens/<domain>/` | L402 cached tokens |
| `~/.lnd/` | lnd data (chain, macaroons, TLS) |
| `~/.lnd/data/chain/bitcoin/<network>/admin.macaroon` | Admin macaroon |
| `~/.lnd/tls.cert` | lnd TLS certificate |
| `~/.lnd-signer/` | Signer lnd data |
| `~/.aperture/aperture.yaml` | Aperture configuration |
| `~/.aperture/aperture.db` | Aperture token database |
| `mcp-server/.env` | MCP server config |

## Ports

| Port | Service | Daemon |
|------|---------|--------|
| 9735 | Lightning P2P | lnd |
| 10009 | gRPC | lnd |
| 8080 | REST | lnd |
| 10012 | gRPC | signer lnd |
| 10013 | REST | signer lnd |
| 8081 | HTTP/L402 | aperture (configurable) |
