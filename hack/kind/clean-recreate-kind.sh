#!/bin/bash
set -e

# Get the directory that this script is in
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=$(cd "${SCRIPT_DIR}/../.." && pwd)

# Delete existing kubanana cluster if it exists
if kind get clusters | grep -q "kubanana"; then
  echo "Deleting existing kubanana cluster..."
  kind delete cluster --name kubanana
fi

# Create a new kubanana cluster
echo "Creating a new kubanana cluster..."
kind create cluster --name kubanana --config "${SCRIPT_DIR}/kind-config.yaml"

# Load just bare images, don't install anything
echo "Building Docker images..."
make docker-build

echo "Loading images into Kind..."
kind load docker-image kubanana-controller:latest --name kubanana
kind load docker-image busybox --name kubanana
kind load docker-image nginx --name kubanana

echo "Kubanana cluster recreated successfully."