#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")
ROOT_DIR=$(cd "${SCRIPT_DIR}/../.." && pwd)

if ! kind get clusters | grep -q kubevent; then
  echo "Creating Kind cluster 'kubevent'..."
  kind create cluster --config "${SCRIPT_DIR}/kind-config.yaml"
else
  echo "Kind cluster 'kubevent' already exists."
fi

echo "Building Docker images..."
make docker-build

echo "Loading images into Kind..."
kind load docker-image kubevent-controller:latest --name kubevent
kind load docker-image busybox --name kubevent
kind load docker-image nginx --name kubevent

echo "Installing CRD..."
kubectl apply -f "${ROOT_DIR}/deploy/crds/kubevent.roshanbhatia.com_eventtriggeredjobs.yaml"

echo "Setting up namespace and RBAC..."
kubectl apply -f "${ROOT_DIR}/deploy/manifests/rbac.yaml"

echo "Deploying controller and operator..."
kubectl apply -f "${ROOT_DIR}/deploy/manifests/deployment.yaml"