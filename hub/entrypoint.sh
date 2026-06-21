#!/bin/sh
set -e

CONFIG=/etc/couchdev/config.json

if [ -n "$TOKEN" ]; then
    HASH=$(printf '%s' "$TOKEN" | sha256sum | cut -d' ' -f1)
    mkdir -p /etc/couchdev
    printf '{"listen_addr":"%s","token_hash":"%s","db_path":"%s"}\n' \
        "${LISTEN_ADDR:-:8080}" "$HASH" "${DB_PATH:-/mnt/couchdev/hub.db}" \
        > "$CONFIG"
fi

exec couchdev serve --config "$CONFIG"
