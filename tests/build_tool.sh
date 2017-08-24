#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the build script for the Monax stack. It will
# build the tool into docker containers in a reliable and
# predictable manner.

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally

# ----------------------------------------------------------
# USAGE

# build_tool.sh

# ----------------------------------------------------------

IMAGE=quay.io/monax/db
VERSION_REGEX="^v[0-9]+\.[0-9]+\.[0-9]+$"

set -e

if [ "$JENKINS_URL" ] || [ "$CIRCLE_BRANCH" ] || [ "$CIRCLE_TAG" ]
then
  REPO=`pwd`
  CI="true"
else
  REPO=$GOPATH/src/github.com/hyperledger/burrow
fi

version=$(cat $REPO/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
tag=$(git tag --points-at HEAD)

# Only label a build as a release version when the commit is tagged
if [[ ${tag} =~ ${VERSION_REGEX} ]] ; then
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


# Build
mkdir -p $REPO/target/docker
docker build -t $IMAGE:build $REPO
docker run --rm --entrypoint cat $IMAGE:build /usr/local/bin/burrow > $REPO/target/docker/burrow.dockerartefact
docker run --rm --entrypoint cat $IMAGE:build /usr/local/bin/burrow-client > $REPO/target/docker/burrow-client.dockerartefact
docker build -t $IMAGE:$version -f Dockerfile.deploy $REPO

# If provided, tag the image with the label provided
if [ "$1" ]
then
  docker tag $IMAGE:$version $IMAGE:$1
  docker rmi $IMAGE:$version
fi

# Cleanup
rm $REPO/target/docker/burrow.dockerartefact
rm $REPO/target/docker/burrow-client.dockerartefact

# Remove build image so we don't push it when we push all tags
docker rmi -f $IMAGE:build
