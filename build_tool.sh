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
VERSION_REGEX="^v[0-9]+\.[0-9]+\.[0-9]+$"

version=$(go run ./util/version/cmd/main.go)
tag=$(git tag --points-at HEAD)

if [[ ${tag} =~ ${VERSION_REGEX} ]] ; then
    # Only label a build as a release version when the commit is tagged
    echo "Building release version (tagged $tag)..."
    # Fail noisily when trying to build a release version that does not match code tag
    if [[ ! ${tag} = "v$version" ]]; then
        echo "Build failure: version tag $tag does not match version/version.go version $version"
        exit 1
    fi
else
    date=$(date +"%Y%m%d")
    commit=$(git rev-parse --short HEAD)
    version="$version-dev-$date-$commit"
    echo "Building non-release version $version..."
fi

if [[ "$1" ]] ; then
    # If argument provided, use it as the version tag
    echo "Overriding detected version $version and tagging image as $1"
    version="$1"
fi

docker build -t ${IMAGE}:${version} .

