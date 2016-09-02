#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the build script for eris-db. It will build the
# tool into docker containers in a reliable and predicatable
# manner.

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally

# ----------------------------------------------------------
# USAGE

# build_tool.sh

# ----------------------------------------------------------
# Set defaults
set -e
start=`pwd`
if [ "$CIRCLE_BRANCH" ]
then
  repo=`pwd`
else
  base=github.com/eris-ltd/eris-db
  repo=$GOPATH/src/$base
fi
branch=${CIRCLE_BRANCH:=master}
branch=${branch/-/_}

release_min=$(cat $repo/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
release_maj=$(echo $release_min | cut -d . -f 1-2)

image_base=quay.io/eris/erisdb

cd $repo

if [[ "$branch" = "master" ]]
then
  docker build -t $image_base:latest $repo
  docker tag $image_base:latest $image_base:$release_maj
  docker tag $image_base:latest $image_base:$release_min
else
  docker build -t $image_base:$release_min $repo
fi
test_exit=$?
cd $start
exit $test_exit
