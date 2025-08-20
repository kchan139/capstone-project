#!/bin/bash
set -e

PROJECT_ROOT="$(dirname "$0")/../.."
TERRAFORM_DIR="$PROJECT_ROOT/terraform"
ANSIBLE_DIR="$PROJECT_ROOT/ansible"
INVENTORY_FILE="$ANSIBLE_DIR/inventory.ini"

source "$ANSIBLE_DIR/scripts/.env"

# Generate inventory.ini
echo "[servers]" > "$INVENTORY_FILE"
tofu -chdir="$TERRAFORM_DIR" output -json capstone_droplet_ip \
  | jq -r --arg port "$SSH_PORT" '. + " ansible_port=" + $port' \
  >> "$INVENTORY_FILE"

echo "Generated inventory:"
cat "$INVENTORY_FILE"
echo

# Run SSH key generation playbook
ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook \
    -i "$INVENTORY_FILE" \
    --private-key ~/.ssh/id_ed25519 \
    -u "$USER" \
    -e ssh_port=$SSH_PORT \
    "$ANSIBLE_DIR/generate-ssh-keys.yml" \
    --ask-vault-pass \
    --ask-become-pass
