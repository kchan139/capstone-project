#!/bin/bash
set -e

PROJECT_ROOT="$(dirname "$0")"
source "$PROJECT_ROOT/.env"

ssh -L "$LOCAL_PORT":127.0.0.1:"$REMOTE_PORT" "$USERNAME"@"$SERVER_IP" -p "$SSH_PORT"
