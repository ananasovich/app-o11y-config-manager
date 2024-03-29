#!/usr/bin/env bash
set -eufo pipefail

CLUSTER_NAME="app-o11y-config-manager"
IMAGE="$1"

if [[ "$IMAGE" == "" ]]; then
  echo "usage: push_image.sh <image_name:tag>"
  exit 1
fi

k3d image import "${IMAGE}" -c "${CLUSTER_NAME}"