#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <component> [--follow]"
  echo "  component: 'controller'"
  echo "  --follow: Optional flag to follow logs"
  exit 1
fi

COMPONENT=$1
FOLLOW_FLAG=""

if [ $# -eq 2 ] && [ "$2" == "--follow" ]; then
  FOLLOW_FLAG="-f"
fi

if [ "$COMPONENT" == "controller" ]; then
  POD=$(kubectl -n kubevent-system get pods -l app=kubevent-controller -o name | head -n 1)
else
  echo "Unknown component: $COMPONENT. Must be 'controller'."
  exit 1
fi

if [ -z "$POD" ]; then
  echo "No pod found for $COMPONENT."
  exit 1
fi

echo "Viewing logs for $POD..."
kubectl -n kubevent-system logs "$POD" $FOLLOW_FLAG