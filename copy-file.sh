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

DEST="$PROJECT_ROOT/tmp-local/"

if [ "$#" -ne 1 ]; then
  echo "Copy a remote file into $DEST"
  echo "Usage: $0 <remote-path>"
  echo "Example: $0 /home/user/example.txt"
  exit 1
fi

mkdir -p "$DEST"
REMOTE_PATH="$1"

echo "Copying $1 from $SERVER_IP"
scp -P "$SSH_PORT" "$USERNAME@$SERVER_IP:$REMOTE_PATH" "$DEST"
echo "File copied to: $DEST"
