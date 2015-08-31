#!/bin/sh

release_maj="0.10"
release_min="0.10.2"

start=`pwd`
branch=${ERISDB_BUILD_BRANCH:=master}
base=github.com/eris-ltd/eris-db
repo=$GOPATH/src/$base

cd $repo

# if [ "$DEV" != "true" ]; then
# 	git checkout $branch
# 	git pull origin
# fi

if [ "$ERISDB_BUILD_BRANCH" = "master" ]; then
  docker build -t eris/erisdb:latest -f DOCKER/Dockerfile .
  docker tag -f eris/erisdb:latest eris/erisdb:$release_maj
  docker tag -f eris/erisdb:latest eris/erisdb:$release_min
else
  docker build -t eris/erisdb:$branch -f DOCKER/Dockerfile .
fi

cd $start
