#!/usr/bin/env bash
# Shared REST call helpers for lnd/litd scripts.
#
# Provides rest_call() and wait_for_rest() functions that handle both
# native and container modes. In container mode, the host-mapped port
# is discovered via `docker port` since container images do not include
# curl.
#
# Required variables (set before sourcing):
#   REST_HOST       - REST API host (default: localhost).
#   REST_PORT       - REST API port (default: 8080).
#   CONTAINER       - Docker container name (empty for native mode).
#   CONTAINER_PORT  - Internal container port to map (default: $REST_PORT).
#
# Usage:
#   source skills/lib/rest.sh
#   rest_call GET "/v1/state"
#   rest_call POST "/v1/initwallet" '{"wallet_password":"..."}'
#   wait_for_rest

# rest_call issues a curl request against the lnd/litd REST API.
# In container mode, it discovers the host-mapped port via docker port.
rest_call() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"

    local host="${REST_HOST:-localhost}"
    local port="${REST_PORT:-8080}"
    local internal_port="${CONTAINER_PORT:-$port}"

    # Container mode: discover the host-mapped port for the container's
    # internal REST port.
    if [ -n "${CONTAINER:-}" ]; then
        host="localhost"
        local mapped
        mapped=$(docker port "$CONTAINER" "$internal_port" 2>/dev/null \
            | head -1 | sed 's/.*://')
        if [ -n "$mapped" ]; then
            port="$mapped"
        fi
    fi

    if [ "$method" = "GET" ]; then
        curl -sk -X GET \
            "https://$host:$port$endpoint" 2>&1
    else
        curl -sk -X POST \
            "https://$host:$port$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" 2>&1
    fi
}

# wait_for_rest polls the REST API until it responds or times out.
wait_for_rest() {
    local label="${1:-REST API}"
    echo "Waiting for $label..."
    for i in {1..30}; do
        if rest_call GET "/v1/state" &>/dev/null; then
            echo "$label is ready."
            return 0
        fi
        sleep 2
        echo "  Waiting... ($i/30)"
    done
    echo "Error: $label did not become available." >&2
    return 1
}
