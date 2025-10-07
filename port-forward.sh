#!/bin/bash
set -e

PROJECT_ROOT="$(dirname "$0")"
source "$PROJECT_ROOT/.env"

for mapping in $PORTS; do
    LOCAL_PORT="${mapping%%:*}"
    REMOTE_PORT="${mapping##*:}"
    SSH_ARGS+=("-L" "${LOCAL_PORT}:${REMOTE_HOST}:${REMOTE_PORT}")
done

ssh "${SSH_ARGS[@]}" "$USERNAME@$SERVER_IP" -p "$SSH_PORT"
