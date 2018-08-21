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

# Gives RFC 3339 with T instead of space
date=$(date -Iseconds)

docker build --build-arg VERSION=${version}\
 --build-arg VCS_REF=${commit}\
 --build-arg BUILD_DATE=${date}\
 -t ${DOCKER_REPO}:${version} ${REPO}
# Quick smoke test
echo "Emitting version from docker image as smoke test..."
docker run ${DOCKER_REPO}:${version} -v

