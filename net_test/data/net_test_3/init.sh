#! /bin/bash
# this script installs and runs erisdb, because we don't have patience for docker

BRANCH="fixes"

go get -d github.com/monax/burrow
cd $GOPATH/src/github.com/monax/burrow
git fetch origin $BRANCH
git checkout $BRANCH
go install ./cmd/erisdb

erisdb $CHAIN_DIR # should be exposed by docker file
