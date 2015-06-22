#! /bin/sh

CUR=`pwd`

if [ $CUR = "$GOPATH/src/github.com/eris-ltd/erisdb" ]; then

	docker build -t eris-db -f DOCKER/Dockerfile .
else
	docker build -t eris-db -f Dockerfile ..
fi



