#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the build script for the eris stack. It will
# build the tool into docker containers in a reliable and
# predicatable manner.

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally

# ----------------------------------------------------------
# USAGE

# build_tool.sh

# ----------------------------------------------------------

TARGET=eris-db
IMAGE=quay.io/eris/db

set -e

if [ "$JENKINS_URL" ] || [ "$CIRCLE_BRANCH" ]
then
  REPO=`pwd`
  CI="true"
else
  REPO=$GOPATH/src/github.com/eris-ltd/$TARGET
fi

release_min=$(cat $REPO/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
release_maj=$(echo $release_min | cut -d . -f 1-2)

# Build
docker build -t $IMAGE:build $REPO
docker run --rm --entrypoint cat $IMAGE:build /usr/local/bin/$TARGET > $REPO/"$TARGET"_build_artifact
docker run --rm --entrypoint cat $IMAGE:build /usr/local/bin/eris-client > $REPO/eris-client
docker build -t $IMAGE:$release_min -f Dockerfile.deploy $REPO

# If provided, tag the image with the label provided
if [ "$1" ]
then
  docker tag $IMAGE:$release_min $IMAGE:$1
  docker rmi $IMAGE:$release_min
fi

# Cleanup
rm $REPO/"$TARGET"_build_artifact
rm $REPO/eris-client
docker rmi -f $IMAGE:build
