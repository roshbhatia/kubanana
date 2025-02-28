#!/bin/bash
set -e

# Get the directory that this script is in
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=$(cd "${SCRIPT_DIR}/../.." && pwd)

# Install the CRD manually
echo "Installing CRD..."
kubectl apply -f "${ROOT_DIR}/deploy/crds/kubanana.roshanbhatia.com_eventtriggeredjobs.yaml"

echo "CRD installation completed."