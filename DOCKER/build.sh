#!/bin/sh

<<<<<<< HEAD
release_maj="0.10"
release_min="0.10.2"
=======
if [ "$CIRCLE_BRANCH" ]
then
  repo=`pwd`
else
  base=github.com/eris-ltd/eris-db
  repo=$GOPATH/src/$base
fi
branch=${CIRCLE_BRANCH:=master}

release_min=$(cat $repo/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
release_maj=$(echo $release_min | cut -d . -f 1-2)
>>>>>>> fix_versions

start=`pwd`
image_base=quay.io/eris/erisdb

cd $repo

if [ "$branch" = "master" ]; then
  docker build -t $image_base:latest -f DOCKER/Dockerfile .
  docker tag -f $image_base:latest $image_base:$release_maj
  docker tag -f $image_base:latest $image_base:$release_min
else
  docker build -t $image_base:$branch -f DOCKER/Dockerfile .
fi

cd $start
