#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the build script for the monax stack. It will
# build the tool into docker containers in a reliable and
# predicatable manner.

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally

# ----------------------------------------------------------
# USAGE

# build_tool.sh

# ----------------------------------------------------------

TARGET=monax-keys
IMAGE=quay.io/monax/keys

set -e

if [ "$JENKINS_URL" ] || [ "$CIRCLE_BRANCH" ]
then
  REPO=`pwd`
  CI="true"
else
  REPO=$GOPATH/src/github.com/hyperledger/burrow/keys
fi

release_min=$(cat $REPO/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
release_maj=$(echo $release_min | cut -d . -f 1-2)

# Build
docker build -t $IMAGE:build $REPO
docker run --rm --entrypoint cat $IMAGE:build /usr/local/bin/$TARGET > $REPO/"$TARGET"_build_artifact

#-----------------------------------------------------------------------------

docker build -t $IMAGE:$release_min -f Dockerfile.deploy $REPO

# Cleanup
rm $REPO/"$TARGET"_build_artifact

# Extra Tags
if [[ "$branch" = "master" ]]
then
  docker tag -f $IMAGE:$release_min $IMAGE:$release_maj
  docker tag -f $IMAGE:$release_min $IMAGE:latest
fi

if [ "$CIRCLE_BRANCH" ]
then
  docker tag -f $IMAGE:$release_min $IMAGE:latest
fi
