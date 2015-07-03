#!/bin/bash
base=github.com/eris-ltd/eris-db
repo=$GOPATH/src/$base
branch=${ERISDB_BUILD_BRANCH:=dockerfixes}
start=`pwd`

cd $repo
if [ "$DEV" != "true" ]; then 
	git checkout $branch
	git pull origin
fi

docker build -t eris/erisdb:0.10 -f DOCKER/Dockerfile .

cd $start
