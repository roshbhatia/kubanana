#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


echo "Building Docker images..."
make docker-build

echo "Loading images into Kind..."
kind load docker-image kubevent-controller:latest --name kubevent

echo "Restarting deployments to pick up new images..."
kubectl -n kubevent-system rollout restart deployment/kubevent-controller

echo "Waiting for deployments to be ready..."
kubectl -n kubevent-system rollout status deployment/kubevent-controller --timeout=60s

echo "Redeployment complete."