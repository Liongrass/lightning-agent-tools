#!/usr/bin/env bash
# Shared config generation functions for litd and lnd containers.
#
# These functions take a config template and produce a runtime config
# by substituting network, debug level, node alias, UI password, and
# appending extra arguments as config-file lines.
#
# Usage:
#   source skills/lib/config-gen.sh
#   generate_litd_config <template> <output> [network] [debug] [alias] [uipass] [extra_args] [bitcoind_opts]
#   generate_lnd_config  <template> <output> [network] [debug] [extra_args] [bitcoind_opts]

# Default bitcoind connection values used when switching to regtest.
# Uses the container name (litd-bitcoind) rather than the compose service name
# (bitcoind) so that containers joining via `docker network connect` can
# resolve the host — service aliases only work within the originating compose.
_BITCOIND_RPCHOST="${BITCOIND_RPCHOST:-litd-bitcoind:18443}"
_BITCOIND_RPCUSER="${BITCOIND_RPCUSER:-devuser}"
_BITCOIND_RPCPASS="${BITCOIND_RPCPASS:-devpass}"
_BITCOIND_ZMQPUBRAWBLOCK="${BITCOIND_ZMQPUBRAWBLOCK:-tcp://litd-bitcoind:28332}"
_BITCOIND_ZMQPUBRAWTX="${BITCOIND_ZMQPUBRAWTX:-tcp://litd-bitcoind:28333}"

# generate_litd_config produces a litd runtime config from a template.
#
# It replaces the network boolean flag, debug level, node alias, and
# UI password in-place, then appends any extra arguments as config lines.
# Extra arguments are converted from CLI format (--lnd.flag=value) to
# config file format (lnd.flag=value). Boolean flags without a value
# get =true appended.
#
# When network=regtest, the function automatically swaps
# lnd.bitcoin.node=neutrino to lnd.bitcoin.node=bitcoind, comments out
# neutrino and fee.url lines, and appends bitcoind connection defaults.
#
# Arguments:
#   $1 - template  Path to the .conf.template file.
#   $2 - output    Path to write the generated config.
#   $3 - network   Bitcoin network (default: testnet).
#   $4 - debug     Debug level string (default: info).
#   $5 - alias     Node alias (default: litd-agent).
#   $6 - uipass    litd UI password (default: empty, keeps template value).
#   $7 - extra     Space-separated extra CLI flags to append.
#   $8 - bitcoind  Bitcoind connection overrides (host:user:pass). Optional.
generate_litd_config() {
    local template="$1"
    local output="$2"
    local network="${3:-testnet}"
    local debug_level="${4:-info}"
    local alias="${5:-litd-agent}"
    local ui_password="${6:-}"
    local extra_args="${7:-}"
    local bitcoind_opts="${8:-}"

    cp "$template" "$output"

    # Replace the litd-level network flag: network=<old> -> network=<new>.
    # litd's global network= flag handles setting the correct lnd.bitcoin.*
    # flag internally, avoiding the lnd default config conflict.
    sed -i.bak -E \
        's/^network=(testnet|mainnet|signet|regtest)/network='"$network"'/' \
        "$output"
    rm -f "$output.bak"

    # Replace the debug level.
    sed -i.bak 's/^lnd\.debuglevel=.*/lnd.debuglevel='"$debug_level"'/' "$output"
    rm -f "$output.bak"

    # Replace the node alias.
    sed -i.bak 's/^lnd\.alias=.*/lnd.alias='"$alias"'/' "$output"
    rm -f "$output.bak"

    # Set UI password if provided.
    if [ -n "$ui_password" ]; then
        if grep -q '^uipassword=' "$output"; then
            sed -i.bak 's/^uipassword=.*/uipassword='"$ui_password"'/' "$output"
            rm -f "$output.bak"
        fi
    fi

    # Regtest: swap neutrino backend for bitcoind.
    if [ "$network" = "regtest" ]; then
        _apply_regtest_bitcoind_litd "$output" "$bitcoind_opts"
    fi

    # Append extra args as config-file lines.
    _append_extra_args "$output" "$extra_args"
}

# generate_lnd_config produces a standalone lnd runtime config from a
# template. Same pattern as generate_litd_config but for lnd-style
# configs that use unprefixed keys (bitcoin.active, debuglevel, etc.).
#
# When network=regtest, the function automatically swaps
# bitcoin.node=neutrino to bitcoin.node=bitcoind, comments out neutrino
# and fee.url lines, and appends bitcoind connection defaults.
#
# Arguments:
#   $1 - template  Path to the .conf.template file.
#   $2 - output    Path to write the generated config.
#   $3 - network   Bitcoin network (default: testnet).
#   $4 - debug     Debug level string (default: info).
#   $5 - extra     Space-separated extra CLI flags to append.
#   $6 - bitcoind  Bitcoind connection overrides (host:user:pass). Optional.
generate_lnd_config() {
    local template="$1"
    local output="$2"
    local network="${3:-testnet}"
    local debug_level="${4:-info}"
    local extra_args="${5:-}"
    local bitcoind_opts="${6:-}"

    cp "$template" "$output"

    # Replace the network boolean: bitcoin.<old>=true -> bitcoin.<new>=true.
    # Use -E for extended regex (portable across macOS and Linux).
    sed -i.bak -E \
        's/^bitcoin\.(testnet|mainnet|signet|regtest)=true/bitcoin.'"$network"'=true/' \
        "$output"
    rm -f "$output.bak"

    # Replace the debug level.
    sed -i.bak 's/^debuglevel=.*/debuglevel='"$debug_level"'/' "$output"
    rm -f "$output.bak"

    # Regtest: swap neutrino backend for bitcoind.
    if [ "$network" = "regtest" ]; then
        _apply_regtest_bitcoind_lnd "$output" "$bitcoind_opts"
    fi

    # Append extra args as config-file lines.
    _append_extra_args "$output" "$extra_args"
}

# _apply_regtest_bitcoind_lnd rewrites an lnd config (unprefixed keys) to use
# bitcoind instead of neutrino. Called automatically when network=regtest.
_apply_regtest_bitcoind_lnd() {
    local output="$1"
    local bitcoind_opts="$2"

    # Parse optional overrides (host:user:pass).
    local rpchost="$_BITCOIND_RPCHOST"
    local rpcuser="$_BITCOIND_RPCUSER"
    local rpcpass="$_BITCOIND_RPCPASS"
    local zmqblock="$_BITCOIND_ZMQPUBRAWBLOCK"
    local zmqtx="$_BITCOIND_ZMQPUBRAWTX"
    if [ -n "$bitcoind_opts" ]; then
        IFS=':' read -r _h _u _p <<< "$bitcoind_opts"
        [ -n "$_h" ] && rpchost="$_h"
        [ -n "$_u" ] && rpcuser="$_u"
        [ -n "$_p" ] && rpcpass="$_p"
    fi

    # Swap neutrino for bitcoind.
    sed -i.bak 's/^bitcoin\.node=neutrino/bitcoin.node=bitcoind/' "$output"
    rm -f "$output.bak"

    # Comment out neutrino lines.
    sed -i.bak 's/^neutrino\./#neutrino./' "$output"
    rm -f "$output.bak"

    # Comment out fee.url line (not needed for regtest).
    sed -i.bak 's/^fee\.url=/#fee.url=/' "$output"
    rm -f "$output.bak"

    # Append bitcoind connection settings if not already present.
    if ! grep -q '^bitcoind\.rpchost=' "$output"; then
        echo "" >> "$output"
        echo "# Bitcoind connection settings (regtest)." >> "$output"
        echo "bitcoind.rpchost=$rpchost" >> "$output"
        echo "bitcoind.rpcuser=$rpcuser" >> "$output"
        echo "bitcoind.rpcpass=$rpcpass" >> "$output"
        echo "bitcoind.zmqpubrawblock=$zmqblock" >> "$output"
        echo "bitcoind.zmqpubrawtx=$zmqtx" >> "$output"
    fi
}

# _apply_regtest_bitcoind_litd rewrites a litd config (lnd.-prefixed keys) to
# use bitcoind instead of neutrino. Called automatically when network=regtest.
_apply_regtest_bitcoind_litd() {
    local output="$1"
    local bitcoind_opts="$2"

    # Parse optional overrides (host:user:pass).
    local rpchost="$_BITCOIND_RPCHOST"
    local rpcuser="$_BITCOIND_RPCUSER"
    local rpcpass="$_BITCOIND_RPCPASS"
    local zmqblock="$_BITCOIND_ZMQPUBRAWBLOCK"
    local zmqtx="$_BITCOIND_ZMQPUBRAWTX"
    if [ -n "$bitcoind_opts" ]; then
        IFS=':' read -r _h _u _p <<< "$bitcoind_opts"
        [ -n "$_h" ] && rpchost="$_h"
        [ -n "$_u" ] && rpcuser="$_u"
        [ -n "$_p" ] && rpcpass="$_p"
    fi

    # Swap neutrino for bitcoind.
    sed -i.bak 's/^lnd\.bitcoin\.node=neutrino/lnd.bitcoin.node=bitcoind/' "$output"
    rm -f "$output.bak"

    # Comment out lnd.neutrino lines.
    sed -i.bak 's/^lnd\.neutrino\./#lnd.neutrino./' "$output"
    rm -f "$output.bak"

    # Comment out lnd.fee.url line (not needed for regtest).
    sed -i.bak 's/^lnd\.fee\.url=/#lnd.fee.url=/' "$output"
    rm -f "$output.bak"

    # Append bitcoind connection settings if not already present.
    if ! grep -q '^lnd\.bitcoind\.rpchost=' "$output"; then
        echo "" >> "$output"
        echo "# Bitcoind connection settings (regtest)." >> "$output"
        echo "lnd.bitcoind.rpchost=$rpchost" >> "$output"
        echo "lnd.bitcoind.rpcuser=$rpcuser" >> "$output"
        echo "lnd.bitcoind.rpcpass=$rpcpass" >> "$output"
        echo "lnd.bitcoind.zmqpubrawblock=$zmqblock" >> "$output"
        echo "lnd.bitcoind.zmqpubrawtx=$zmqtx" >> "$output"
    fi
}

# _append_extra_args converts CLI-style flags to config-file lines and
# appends them to the output file. Each --flag=value becomes flag=value;
# bare --flag becomes flag=true.
_append_extra_args() {
    local output="$1"
    local extra_args="$2"

    if [ -z "$extra_args" ]; then
        return
    fi

    echo "" >> "$output"
    echo "# Profile/CLI extra arguments." >> "$output"
    for arg in $extra_args; do
        # Strip leading dashes.
        local line="${arg#--}"
        line="${line#-}"

        # Boolean flags without =value get =true.
        if [[ "$line" != *=* ]]; then
            line="${line}=true"
        fi

        echo "$line" >> "$output"
    done
}
