#!/usr/bin/env bash
# Shared config generation functions for litd and lnd containers.
#
# These functions take a config template and produce a runtime config
# by substituting network, debug level, node alias, UI password, and
# appending extra arguments as config-file lines.
#
# Usage:
#   source skills/lib/config-gen.sh
#   generate_litd_config <template> <output> [network] [debug] [alias] [uipass] [extra_args]
#   generate_lnd_config  <template> <output> [network] [debug] [extra_args]

# generate_litd_config produces a litd runtime config from a template.
#
# It replaces the network boolean flag, debug level, node alias, and
# UI password in-place, then appends any extra arguments as config lines.
# Extra arguments are converted from CLI format (--lnd.flag=value) to
# config file format (lnd.flag=value). Boolean flags without a value
# get =true appended.
#
# Arguments:
#   $1 - template  Path to the .conf.template file.
#   $2 - output    Path to write the generated config.
#   $3 - network   Bitcoin network (default: testnet).
#   $4 - debug     Debug level string (default: info).
#   $5 - alias     Node alias (default: litd-agent).
#   $6 - uipass    litd UI password (default: empty, keeps template value).
#   $7 - extra     Space-separated extra CLI flags to append.
generate_litd_config() {
    local template="$1"
    local output="$2"
    local network="${3:-testnet}"
    local debug_level="${4:-info}"
    local alias="${5:-litd-agent}"
    local ui_password="${6:-}"
    local extra_args="${7:-}"

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

    # Append extra args as config-file lines.
    _append_extra_args "$output" "$extra_args"
}

# generate_lnd_config produces a standalone lnd runtime config from a
# template. Same pattern as generate_litd_config but for lnd-style
# configs that use unprefixed keys (bitcoin.active, debuglevel, etc.).
#
# Arguments:
#   $1 - template  Path to the .conf.template file.
#   $2 - output    Path to write the generated config.
#   $3 - network   Bitcoin network (default: testnet).
#   $4 - debug     Debug level string (default: info).
#   $5 - extra     Space-separated extra CLI flags to append.
generate_lnd_config() {
    local template="$1"
    local output="$2"
    local network="${3:-testnet}"
    local debug_level="${4:-info}"
    local extra_args="${5:-}"

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

    # Append extra args as config-file lines.
    _append_extra_args "$output" "$extra_args"
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
