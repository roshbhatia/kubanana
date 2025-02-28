#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")
ROOT_DIR=$(cd "${SCRIPT_DIR}/../.." && pwd)

if ! kubectl get namespace test-kubanana &>/dev/null; then
  echo "Creating test namespace..."
  kubectl create namespace test-kubanana
fi

echo "Applying EventTriggeredJob..."
kubectl apply -f "${ROOT_DIR}/deploy/samples/example-job-template.yaml"

echo "Creating a test pod to trigger an event..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: default
  labels:
    app: myapp
spec:
  containers:
  - name: alpine
    image: alpine:latest
    command: ["sleep", "30"]
EOF

echo "Test pod created. Waiting for 5 seconds..."
sleep 5

echo "Deleting the test pod to trigger a DELETE event..."
kubectl delete pod test-pod

echo "Checking for jobs created by the controller..."
sleep 5
kubectl get jobs

echo "Test completed. Check the logs to see if the job was triggered."
echo "Run: ./hack/kind/view-logs.sh controller"