#!/bin/bash
set -e

# Get the directory that this script is in
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=$(cd "${SCRIPT_DIR}/../.." && pwd)

# Delete existing kubevent cluster if it exists
if kind get clusters | grep -q "kubevent"; then
  echo "Deleting existing kubevent cluster..."
  kind delete cluster --name kubevent
fi

# Create a new kubevent cluster
echo "Creating a new kubevent cluster..."
kind create cluster --name kubevent --config "${SCRIPT_DIR}/kind-config.yaml"

# Load just bare images, don't install anything
echo "Building Docker images..."
make docker-build

echo "Loading images into Kind..."
kind load docker-image kubevent-controller:latest --name kubevent
kind load docker-image busybox --name kubevent
kind load docker-image nginx --name kubevent

echo "Kubevent cluster recreated successfully."