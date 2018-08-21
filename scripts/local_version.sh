#!/usr/bin/env bash
#
# Script that outputs a version identifier based on the in-code version of
# Burrow combined with date and git commit or the git tag if tag is a version.
#
# If working directory is checked out at a version tag then checks that the tag
# matches the in-code version and fails if it does not.
#
set -e

REPO=${REPO:-"$GOPATH/src/github.com/hyperledger/burrow"}
VERSION_REGEX="^v[0-9]+\.[0-9]+\.[0-9]+$"


version=$(go run "$REPO/project/cmd/version/main.go")
tag=$(git tag --points-at HEAD)

function log() {
    echo "$*" >> /dev/stderr
}

# Same as specified RFC3339 but contains the T
date=$(date -Idate)
commit=$(git rev-parse --short HEAD)

if [[ ${tag} =~ ${VERSION_REGEX} ]] ; then
    # Only label a build as a release version when the commit is tagged
    log "Building release version (tagged $tag)..."
    # Fail noisily when trying to build a release version that does not match code tag
    if [[ ! ${tag} = "v$version" ]]; then
        log "Build failure: version tag $tag does not match version/version.go version $version"
        exit 1
    fi
else
    version="$version-dev-$date-$commit"
    log "Building non-release version $version..."
fi

echo ${version}