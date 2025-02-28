#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")
ROOT_DIR=$(cd "${SCRIPT_DIR}/../.." && pwd)

# Call the clean-recreate-kind.sh script to create a clean cluster and load images
${SCRIPT_DIR}/clean-recreate-kind.sh

# Now manually install components
echo "Installing CRD..."
kubectl apply -f "${ROOT_DIR}/deploy/crds/kubanana.roshanbhatia.com_eventtriggeredjobs.yaml"

echo "Setting up namespace and RBAC..."
kubectl apply -f "${ROOT_DIR}/deploy/manifests/rbac.yaml"

echo "Deploying controller and operator..."
kubectl apply -f "${ROOT_DIR}/deploy/manifests/deployment.yaml"