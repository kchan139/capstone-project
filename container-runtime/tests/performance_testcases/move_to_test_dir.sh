#!/bin/bash

if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Usage: ./deploy.sh <program> <test_dir>"
    echo "Example: ./deploy.sh p1 test_1"
    exit 1
fi

PROGRAM=$1
DEST="/home/phiung/cont-test/rootfs/home/khoatran/$2"

mv $PROGRAM $DEST
echo "Deployed $PROGRAM to $DEST"