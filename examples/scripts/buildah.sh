#!/usr/bin/env bash

# Script to build a container using Buildah
# NOTE sudo privileges to create and remove /var/lib/mycontainer directory
# Usage:
# sh buildah.sh REGISTRY_ADDRESS/USER IMAGE_NAME
# Example:
# sudo sh buildah.sh quay.io/dsalerno buildhatest
mkdir /var/lib/mycontainer
podman run -v ../../builder/testdata:/build:z -v /var/lib/mycontainer:/var/lib/containers:Z --device /dev/fuse:rw --security-opt seccomp=unconfined --security-opt apparmor=unconfined quay.io/buildah/stable buildah -t $2 bud -f /build/Dockerfile.buildah .
podman run -v ../../builder/testdata:/build:z -v /var/lib/mycontainer:/var/lib/containers:Z --device /dev/fuse:rw --security-opt seccomp=unconfined --security-opt apparmor=unconfined quay.io/buildah/stable buildah push  $2 $1
rm -rf /var/lib/mycontainer