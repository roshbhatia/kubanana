#!/bin/bash
set -e

# Get the directory that this script is in
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Set base directory
BASE_DIR="$SCRIPT_DIR/../.."

# Recreate the kind cluster to start fresh (includes building & loading images)
echo "Recreating kind cluster..."
"$SCRIPT_DIR"/clean-recreate-kind.sh

# Install CRDs manually first
echo "Installing CRDs manually..."
"$SCRIPT_DIR"/install-crds.sh

# Install Helm chart (with installCRDs set to false)
echo "Installing Helm chart..."
helm upgrade --install kubanana "$BASE_DIR"/charts/kubanana \
  --create-namespace \
  --namespace kubanana-system \
  --set deployment.image.tag=latest \
  --set installCRDs=false

# Wait for the controller to be ready
echo "Waiting for controller deployment to be ready..."
kubectl wait --for=condition=available --timeout=60s deployment/kubanana-controller -n kubanana-system

# Create a test EventTriggeredJob resource
echo "Creating test EventTriggeredJob resource..."
kubectl apply -f "$BASE_DIR"/deploy/samples/example-job-template.yaml

# Create a test pod to trigger the job
echo "Creating test pod to trigger EventTriggeredJob..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  labels:
    app: myapp
spec:
  containers:
  - name: busybox
    image: busybox
    command: ["sleep", "300"]
EOF

# Wait for the job to be created
echo "Checking for jobs created by the controller..."
sleep 5

# List jobs
kubectl get jobs

echo "Test completed."