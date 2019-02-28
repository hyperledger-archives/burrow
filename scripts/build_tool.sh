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

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Grab date, commit, version
. "$script_dir/local_version.sh" > /dev/null

DOCKER_REPO=${DOCKER_REPO:-"hyperledger/burrow"}
REPO=${REPO:-"$GOPATH/src/github.com/hyperledger/burrow"}

function log() {
    echo "$*" >> /dev/stderr
}

if [[ "$1" ]] ; then
    # If argument provided, use it as the version tag
    log "Overriding detected version $version and tagging image as $1"
    version="$1"
fi

# Expiry is intended for dev images, if we want more persistent Burrow images on quay.io we should remove this...
docker build \
  --label quay.expires-after=24w\
  --label org.label-schema.version=${version}\
  --label org.label-schema.vcs-ref=${commit}\
  --label org.label-schema.build-date=${date}\
  -t ${DOCKER_REPO}:${version} ${REPO}

# Quick smoke test
echo "Emitting version from docker image as smoke test..."
docker run ${DOCKER_REPO}:${version} -v

