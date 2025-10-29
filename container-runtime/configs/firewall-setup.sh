#!/bin/bash
# MRUNC Container Firewall Setup Script
# Usage: ./firewall-setup.sh <veth-interface> <container-ip>

# ** NOTE: This firewall rule is very relaxed and only suitable for dev environment **

set -e

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <veth-interface> <container-ip>"
    exit 1
fi

VETH_HOST="$1"
CONTAINER_IP="$2"

echo "Setting up firewall rules for container..."
echo "  Interface: $VETH_HOST"
echo "  Container IP: $CONTAINER_IP"

# Enable IP forwarding
echo 1 > /proc/sys/net/ipv4/ip_forward

# Setup NAT
iptables -t nat -A POSTROUTING -s ${CONTAINER_IP}/32 -j MASQUERADE

# Allow forwarding for this container
iptables -A FORWARD -i ${VETH_HOST} -j ACCEPT
iptables -A FORWARD -o ${VETH_HOST} -m state --state RELATED,ESTABLISHED -j ACCEPT

echo "  Firewall rules applied successfully"
echo ""
echo "To remove these rules later:"
echo "  iptables -t nat -D POSTROUTING -s ${CONTAINER_IP}/32 -j MASQUERADE"
echo "  iptables -D FORWARD -i ${VETH_HOST} -j ACCEPT"
echo "  iptables -D FORWARD -o ${VETH_HOST} -m state --state RELATED,ESTABLISHED -j ACCEPT"
