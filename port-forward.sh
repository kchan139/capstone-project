#!/bin/bash
set -e

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do
  DIR="$(cd -P "$(dirname "$SOURCE")" && pwd)"
  SOURCE="$(readlink "$SOURCE")"
  [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
done
PROJECT_ROOT="$(cd -P "$(dirname "$SOURCE")" && pwd)"

if [ -f "$PROJECT_ROOT/.env" ]; then
  source "$PROJECT_ROOT/.env"
else
  echo ".env not found in $PROJECT_ROOT" >&2
  exit 1
fi

for mapping in $PORTS; do
    LOCAL_PORT="${mapping%%:*}"
    REMOTE_PORT="${mapping##*:}"
    SSH_ARGS+=("-L" "${LOCAL_PORT}:${REMOTE_HOST}:${REMOTE_PORT}")
done

SSH_ARGS+=("-Y")

ssh "${SSH_ARGS[@]}" "$USERNAME@$SERVER_IP" -p "$SSH_PORT"
