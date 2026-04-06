#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RUNTIME_DIR="$ROOT_DIR/container-runtime"
BUNDLE_DIR="$(mktemp -d)"
CONTAINER_ID="ci-test-container"

cleanup() {
  sudo ip link delete veth-ci-test 2>/dev/null || true
  sudo iptables -t nat -D POSTROUTING -s 10.0.0.10/32 -j MASQUERADE 2>/dev/null || true
  sudo iptables -D FORWARD -i veth-ci-test -j ACCEPT 2>/dev/null || true
  sudo iptables -D FORWARD -o veth-ci-test -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || true
  sudo rm -rf "/run/mrunc/$CONTAINER_ID" 2>/dev/null || true
  rm -rf "$BUNDLE_DIR"
}
trap cleanup EXIT

cd "$RUNTIME_DIR"
cp ./configs/examples/ci-test.json "$BUNDLE_DIR/config.json"

echo "Building mrunc..."
make build

echo "Initializing rootfs..."
sudo ./bin/mrunc init

test -f /var/lib/mrunc/images/ubuntu/etc/os-release

echo "Running privileged container smoke test..."
sudo timeout 180 ./bin/mrunc run --bundle "$BUNDLE_DIR" "$CONTAINER_ID"

echo "Privileged smoke test completed."
