#!/bin/bash

set -e

PROJECT_ROOT="$(dirname "$0")/../.."
TERRAFORM_DIR="$PROJECT_ROOT/terraform"
ANSIBLE_DIR="$PROJECT_ROOT/ansible"
INVENTORY_FILE="$ANSIBLE_DIR/inventory.ini"

# Generate inventory.ini
echo "[servers]" > "$INVENTORY_FILE"
tofu -chdir="$TERRAFORM_DIR" output -json capstone_droplet_ip \
  | jq \
  >> "$INVENTORY_FILE"

echo "Generated inventory:"
cat "$INVENTORY_FILE"
echo

# Run Ansible playbook using the generated inventory
ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook \
    -i "$INVENTORY_FILE" \
    --private-key ~/.ssh/id_ed25519 \
    -u root \
    "$ANSIBLE_DIR/playbook.yml" \
    --ask-vault-pass
