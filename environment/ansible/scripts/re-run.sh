#!/bin/bash
set -e

PROJECT_ROOT="$(dirname "$0")/../.."
TERRAFORM_DIR="$PROJECT_ROOT/terraform"
ANSIBLE_DIR="$PROJECT_ROOT/ansible"
INVENTORY_FILE="$ANSIBLE_DIR/inventory.ini"

# Prompt for vault password once and store in variable
echo -n "Vault password: "
read -rs VAULT_PASS
echo

# Get SSH port and user from vault using the stored password
SSH_PORT=$(echo "$VAULT_PASS" | ansible-vault view "$ANSIBLE_DIR/vars/secrets.yml" --vault-password-file=/dev/stdin | grep "^ssh_port:" | awk '{print $2}')
USER=$(echo "$VAULT_PASS" | ansible-vault view "$ANSIBLE_DIR/vars/secrets.yml" --vault-password-file=/dev/stdin | grep "^user1:" | awk '{print $2}')

# Generate inventory.ini
echo "[servers]" > "$INVENTORY_FILE"
tofu -chdir="$TERRAFORM_DIR" output -json capstone_droplet_ip \
  | jq -r --arg port "$SSH_PORT" '. + " ansible_port=" + $port' \
  >> "$INVENTORY_FILE"

echo "Generated inventory:"
cat "$INVENTORY_FILE"
echo

# Run Ansible playbook using the generated inventory and stored password
ANSIBLE_HOST_KEY_CHECKING=False echo "$VAULT_PASS" | ansible-playbook \
    -i "$INVENTORY_FILE" \
    --private-key ~/.ssh/id_ed25519 \
    -u "$USER" \
    -e ssh_port="$SSH_PORT" \
    "$ANSIBLE_DIR/playbook.yml" \
    --vault-password-file=/dev/stdin \
    --ask-become-pass
