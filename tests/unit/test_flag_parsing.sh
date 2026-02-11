#!/usr/bin/env bash
# Test flag parsing for all skill scripts.
#
# Tests for every script:
#   1. --help exits 0 and output contains "Usage:"
#   2. Unknown flag exits non-zero (where applicable)
#   3. Default values are correct (via help text or flag parsing output)
#
# No Docker required — uses mocks.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib/helpers.sh"
source "$SCRIPT_DIR/../lib/mocks.sh"

SKILLS_DIR="$REPO_ROOT/skills"

# ──────────────────────────────────────────────────────────────────
# Helper: test --help for a script.
# ──────────────────────────────────────────────────────────────────
test_help() {
    local script="$1"
    local label="$2"

    begin_test "$label: --help exits 0"
    local output
    local code=0
    output=$("$script" --help 2>&1) || code=$?

    # Some scripts that delegate (start-lnd.sh) pass --help through
    # to docker-start.sh, which the mock handles. Accept exit 0.
    assert_eq "$code" "0" "--help should exit 0"
    assert_contains "$output" "Usage" "--help output should contain 'Usage'"
    end_test
}

# ──────────────────────────────────────────────────────────────────
# Helper: test unknown flag for a script.
# ──────────────────────────────────────────────────────────────────
test_unknown_flag() {
    local script="$1"
    local label="$2"

    begin_test "$label: unknown flag exits non-zero"
    local code=0
    "$script" --bogus-flag-xyz >/dev/null 2>&1 || code=$?

    # Should exit non-zero.
    assert_neq "$code" "0" "unknown flag should exit non-zero"
    end_test
}

# ──────────────────────────────────────────────────────────────────
# Helper: test help output contains expected default info.
# ──────────────────────────────────────────────────────────────────
test_help_contains() {
    local script="$1"
    local label="$2"
    local pattern="$3"

    begin_test "$label: help mentions '$pattern'"
    local output
    local code=0
    output=$("$script" --help 2>&1) || code=$?
    assert_contains "$output" "$pattern" "help should mention '$pattern'"
    end_test
}

# ──────────────────────────────────────────────────────────────────
# Set up mocks so scripts don't try to call real Docker.
# ──────────────────────────────────────────────────────────────────
setup_mocks
mock_docker_ps ""

# ══════════════════════════════════════════════════════════════════
# lnd skill scripts
# ══════════════════════════════════════════════════════════════════

print_header "lnd/scripts"

# --- install.sh ---
test_help "$SKILLS_DIR/lnd/scripts/install.sh" "lnd/install.sh"
test_unknown_flag "$SKILLS_DIR/lnd/scripts/install.sh" "lnd/install.sh"
test_help_contains "$SKILLS_DIR/lnd/scripts/install.sh" "lnd/install.sh" "Docker"
test_help_contains "$SKILLS_DIR/lnd/scripts/install.sh" "lnd/install.sh" "source"

# --- docker-start.sh ---
test_help "$SKILLS_DIR/lnd/scripts/docker-start.sh" "lnd/docker-start.sh"
test_unknown_flag "$SKILLS_DIR/lnd/scripts/docker-start.sh" "lnd/docker-start.sh"
test_help_contains "$SKILLS_DIR/lnd/scripts/docker-start.sh" "lnd/docker-start.sh" "watchonly"
test_help_contains "$SKILLS_DIR/lnd/scripts/docker-start.sh" "lnd/docker-start.sh" "regtest"
test_help_contains "$SKILLS_DIR/lnd/scripts/docker-start.sh" "lnd/docker-start.sh" "profile"
test_help_contains "$SKILLS_DIR/lnd/scripts/docker-start.sh" "lnd/docker-start.sh" "testnet"

# --- docker-stop.sh ---
test_help "$SKILLS_DIR/lnd/scripts/docker-stop.sh" "lnd/docker-stop.sh"
test_unknown_flag "$SKILLS_DIR/lnd/scripts/docker-stop.sh" "lnd/docker-stop.sh"
test_help_contains "$SKILLS_DIR/lnd/scripts/docker-stop.sh" "lnd/docker-stop.sh" "clean"

# --- start-lnd.sh ---
# Note: start-lnd.sh delegates to docker-start.sh by default, but --help
# is handled directly.
test_help "$SKILLS_DIR/lnd/scripts/start-lnd.sh" "lnd/start-lnd.sh"

begin_test "lnd/start-lnd.sh: --native --help shows native options"
code=0
output=$("$SKILLS_DIR/lnd/scripts/start-lnd.sh" --native --help 2>&1) || code=$?
assert_contains "$output" "Usage" "native --help should show usage"
end_test

# --- stop-lnd.sh ---
test_help "$SKILLS_DIR/lnd/scripts/stop-lnd.sh" "lnd/stop-lnd.sh"

# --- create-wallet.sh ---
test_help "$SKILLS_DIR/lnd/scripts/create-wallet.sh" "lnd/create-wallet.sh"
test_unknown_flag "$SKILLS_DIR/lnd/scripts/create-wallet.sh" "lnd/create-wallet.sh"
test_help_contains "$SKILLS_DIR/lnd/scripts/create-wallet.sh" "lnd/create-wallet.sh" "watchonly"
test_help_contains "$SKILLS_DIR/lnd/scripts/create-wallet.sh" "lnd/create-wallet.sh" "standalone"
test_help_contains "$SKILLS_DIR/lnd/scripts/create-wallet.sh" "lnd/create-wallet.sh" "testnet"
test_help_contains "$SKILLS_DIR/lnd/scripts/create-wallet.sh" "lnd/create-wallet.sh" "container"
test_help_contains "$SKILLS_DIR/lnd/scripts/create-wallet.sh" "lnd/create-wallet.sh" "8080"

# --- unlock-wallet.sh ---
test_help "$SKILLS_DIR/lnd/scripts/unlock-wallet.sh" "lnd/unlock-wallet.sh"
test_unknown_flag "$SKILLS_DIR/lnd/scripts/unlock-wallet.sh" "lnd/unlock-wallet.sh"
test_help_contains "$SKILLS_DIR/lnd/scripts/unlock-wallet.sh" "lnd/unlock-wallet.sh" "container"
test_help_contains "$SKILLS_DIR/lnd/scripts/unlock-wallet.sh" "lnd/unlock-wallet.sh" "8080"

# --- lncli.sh ---
test_help "$SKILLS_DIR/lnd/scripts/lncli.sh" "lnd/lncli.sh"
test_help_contains "$SKILLS_DIR/lnd/scripts/lncli.sh" "lnd/lncli.sh" "lncli"
test_help_contains "$SKILLS_DIR/lnd/scripts/lncli.sh" "lnd/lncli.sh" "litcli"
test_help_contains "$SKILLS_DIR/lnd/scripts/lncli.sh" "lnd/lncli.sh" "loop"
test_help_contains "$SKILLS_DIR/lnd/scripts/lncli.sh" "lnd/lncli.sh" "tapcli"

# lncli.sh passes unknown flags through to the CLI, so no unknown flag test.
# Instead, test that no-args gives an error.
begin_test "lnd/lncli.sh: no args exits with error"
code=0
output=$("$SKILLS_DIR/lnd/scripts/lncli.sh" 2>&1) || code=$?
assert_neq "$code" "0" "no args should exit non-zero"
assert_contains "$output" "No command" "should say no command specified"
end_test

# --- import-credentials.sh ---
test_help "$SKILLS_DIR/lnd/scripts/import-credentials.sh" "lnd/import-credentials.sh"
test_unknown_flag "$SKILLS_DIR/lnd/scripts/import-credentials.sh" "lnd/import-credentials.sh"

# ══════════════════════════════════════════════════════════════════
# lightning-security-module skill scripts
# ══════════════════════════════════════════════════════════════════

print_header "lightning-security-module/scripts"

# --- install.sh ---
test_help "$SKILLS_DIR/lightning-security-module/scripts/install.sh" "lsm/install.sh"
test_unknown_flag "$SKILLS_DIR/lightning-security-module/scripts/install.sh" "lsm/install.sh"

# --- docker-start.sh ---
test_help "$SKILLS_DIR/lightning-security-module/scripts/docker-start.sh" "lsm/docker-start.sh"
test_unknown_flag "$SKILLS_DIR/lightning-security-module/scripts/docker-start.sh" "lsm/docker-start.sh"
test_help_contains "$SKILLS_DIR/lightning-security-module/scripts/docker-start.sh" "lsm/docker-start.sh" "testnet"

# --- docker-stop.sh ---
test_help "$SKILLS_DIR/lightning-security-module/scripts/docker-stop.sh" "lsm/docker-stop.sh"
test_unknown_flag "$SKILLS_DIR/lightning-security-module/scripts/docker-stop.sh" "lsm/docker-stop.sh"

# --- setup-signer.sh ---
test_help "$SKILLS_DIR/lightning-security-module/scripts/setup-signer.sh" "lsm/setup-signer.sh"
test_unknown_flag "$SKILLS_DIR/lightning-security-module/scripts/setup-signer.sh" "lsm/setup-signer.sh"
test_help_contains "$SKILLS_DIR/lightning-security-module/scripts/setup-signer.sh" "lsm/setup-signer.sh" "testnet"
test_help_contains "$SKILLS_DIR/lightning-security-module/scripts/setup-signer.sh" "lsm/setup-signer.sh" "10012"
test_help_contains "$SKILLS_DIR/lightning-security-module/scripts/setup-signer.sh" "lsm/setup-signer.sh" "10013"

# --- export-credentials.sh ---
test_help "$SKILLS_DIR/lightning-security-module/scripts/export-credentials.sh" "lsm/export-credentials.sh"
test_unknown_flag "$SKILLS_DIR/lightning-security-module/scripts/export-credentials.sh" "lsm/export-credentials.sh"

# --- start-signer.sh ---
test_help "$SKILLS_DIR/lightning-security-module/scripts/start-signer.sh" "lsm/start-signer.sh"

# --- stop-signer.sh ---
test_help "$SKILLS_DIR/lightning-security-module/scripts/stop-signer.sh" "lsm/stop-signer.sh"

# ══════════════════════════════════════════════════════════════════
# macaroon-bakery skill scripts
# ══════════════════════════════════════════════════════════════════

print_header "macaroon-bakery/scripts"

test_help "$SKILLS_DIR/macaroon-bakery/scripts/bake.sh" "bakery/bake.sh"
test_unknown_flag "$SKILLS_DIR/macaroon-bakery/scripts/bake.sh" "bakery/bake.sh"
test_help_contains "$SKILLS_DIR/macaroon-bakery/scripts/bake.sh" "bakery/bake.sh" "pay-only"
test_help_contains "$SKILLS_DIR/macaroon-bakery/scripts/bake.sh" "bakery/bake.sh" "invoice-only"
test_help_contains "$SKILLS_DIR/macaroon-bakery/scripts/bake.sh" "bakery/bake.sh" "read-only"

# ══════════════════════════════════════════════════════════════════
# aperture skill scripts
# ══════════════════════════════════════════════════════════════════

print_header "aperture/scripts"

test_help "$SKILLS_DIR/aperture/scripts/install.sh" "aperture/install.sh"
test_unknown_flag "$SKILLS_DIR/aperture/scripts/install.sh" "aperture/install.sh"

test_help "$SKILLS_DIR/aperture/scripts/setup.sh" "aperture/setup.sh"
test_unknown_flag "$SKILLS_DIR/aperture/scripts/setup.sh" "aperture/setup.sh"
test_help_contains "$SKILLS_DIR/aperture/scripts/setup.sh" "aperture/setup.sh" "8081"
test_help_contains "$SKILLS_DIR/aperture/scripts/setup.sh" "aperture/setup.sh" "insecure"

test_help "$SKILLS_DIR/aperture/scripts/start.sh" "aperture/start.sh"
test_unknown_flag "$SKILLS_DIR/aperture/scripts/start.sh" "aperture/start.sh"

test_help "$SKILLS_DIR/aperture/scripts/stop.sh" "aperture/stop.sh"
test_unknown_flag "$SKILLS_DIR/aperture/scripts/stop.sh" "aperture/stop.sh"

# ══════════════════════════════════════════════════════════════════
# lnget skill scripts
# ══════════════════════════════════════════════════════════════════

print_header "lnget/scripts"

test_help "$SKILLS_DIR/lnget/scripts/install.sh" "lnget/install.sh"
test_unknown_flag "$SKILLS_DIR/lnget/scripts/install.sh" "lnget/install.sh"

# ══════════════════════════════════════════════════════════════════
# lightning-mcp-server skill scripts
# ══════════════════════════════════════════════════════════════════

print_header "lightning-mcp-server/scripts"

test_help "$SKILLS_DIR/lightning-mcp-server/scripts/install.sh" "lightning-mcp-server/install.sh"
test_unknown_flag "$SKILLS_DIR/lightning-mcp-server/scripts/install.sh" "lightning-mcp-server/install.sh"

test_help "$SKILLS_DIR/lightning-mcp-server/scripts/configure.sh" "lightning-mcp-server/configure.sh"
test_unknown_flag "$SKILLS_DIR/lightning-mcp-server/scripts/configure.sh" "lightning-mcp-server/configure.sh"

test_help "$SKILLS_DIR/lightning-mcp-server/scripts/setup-claude-config.sh" "lightning-mcp-server/setup-claude-config.sh"
test_unknown_flag "$SKILLS_DIR/lightning-mcp-server/scripts/setup-claude-config.sh" "lightning-mcp-server/setup-claude-config.sh"

# ══════════════════════════════════════════════════════════════════
# Cleanup and summary
# ══════════════════════════════════════════════════════════════════

teardown_mocks
print_summary
