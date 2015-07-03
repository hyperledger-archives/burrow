#!/bin/bash
base=github.com/eris-ltd/eris-db
repo=$GOPATH/src/$base
branch=${ERISDB_BUILD_BRANCH:=docker}
start=`pwd`

cd $repo
git checkout $branch
git pull origin

docker build -t eris/erisdb:0.10 -f DOCKER/Dockerfile .

cd $start