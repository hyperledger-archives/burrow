#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the build script for the Monax stack. It will
# build the tool into docker containers in a reliable and
# predictable manner.

# ----------------------------------------------------------
# REQUIREMENTS
#
# docker, go, make, and git installed locally

# ----------------------------------------------------------
# USAGE

# build_tool.sh [version tag]

# ----------------------------------------------------------

set -e

IMAGE=${IMAGE:-"quay.io/monax/db"}
REPO=${REPO:-"$GOPATH/src/github.com/hyperledger/burrow"}

function log() {
    echo "$*" >> /dev/stderr
}

version=$("$REPO/scripts/local_version.sh")

if [[ "$1" ]] ; then
    # If argument provided, use it as the version tag
    log "Overriding detected version $version and tagging image as $1"
    version="$1"
fi

docker build -t ${IMAGE}:${version} ${REPO}

