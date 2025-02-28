#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


echo "Building Docker images..."
make docker-build

echo "Loading images into Kind..."
kind load docker-image kubanana-controller:latest --name kubanana

echo "Restarting deployments to pick up new images..."
kubectl -n kubanana-system rollout restart deployment/kubanana-controller

echo "Waiting for deployments to be ready..."
kubectl -n kubanana-system rollout status deployment/kubanana-controller --timeout=60s

echo "Redeployment complete."