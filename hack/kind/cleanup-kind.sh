#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if kind get clusters | grep -q kubanana; then
  echo "Deleting Kind cluster 'kubanana'..."
  kind delete cluster --name kubanana
else
  echo "Kind cluster 'kubanana' does not exist."
fi