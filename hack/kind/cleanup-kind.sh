#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if kind get clusters | grep -q kubevent; then
  echo "Deleting Kind cluster 'kubevent'..."
  kind delete cluster --name kubevent
else
  echo "Kind cluster 'kubevent' does not exist."
fi