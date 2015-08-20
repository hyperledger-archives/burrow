#!/bin/sh

base=github.com/eris-ltd/eris-db
release="0.10"
repo=$GOPATH/src/$base
branch=${ERISDB_BUILD_BRANCH:=develop}
start=`pwd`

cd $repo

if [ "$DEV" != "true" ]; then
	git checkout $branch
	git pull origin
fi

if [ "$ERISDB_BUILD_BRANCH" = "master" ]; then
  docker build -t eris/erisdb:$release -f DOCKER/Dockerfile .
  docker tag eris/erisdb:$release eris/erisdb:latest
else
  docker build -t eris/erisdb:$branch -f DOCKER/Dockerfile .
fi


cd $start
