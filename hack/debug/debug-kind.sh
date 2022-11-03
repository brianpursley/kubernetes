#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

DLV_TARGET=${1:-kube-apiserver}
DLV_PORT=2345

# Build a node-image without optimizations
DBG=1 kind build node-image

# Make a new image (based on the node-image kind just built) containing delve
TEMP_DIR=$(mktemp -d)
cat << EOF > "${TEMP_DIR}"/Dockerfile
FROM kindest/node:latest
ENV DEBIAN_FRONTEND noninteractive
ENV DEBCONF_NOWARNINGS yes
RUN apt-get update && apt-get install -y --no-install-recommends golang delve
EOF
docker build "$TEMP_DIR" -t delve-node-image
rm -rf "$TEMP_DIR"

# Create a cluster using our customized node image that has delve in it
kind create cluster --image=delve-node-image:latest --config <(cat << EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: $DLV_PORT
    hostPort: $DLV_PORT
EOF
)

cleanup() {
  kind delete cluster
  exit 0
}
trap "cleanup" SIGINT

# Delve attach exits after a debug session, so just keep running it for convenience until you press Ctrl+C
while docker exec -it kind-control-plane /bin/sh -c \
  "dlv --listen=:$DLV_PORT --headless=true --api-version=2 attach \$(pidof ${DLV_TARGET})";
do
  echo "Re-running delve attach (Press Ctrl+C to exit)"
done
