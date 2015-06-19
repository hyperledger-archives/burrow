#! /bin/bash

docker build -t eris-db .

# run tendermint 
docker run --name eris-db --volumes-from mintdata -d -p 46656:46656 -p 46657:46657 -e FAST_SYNC=$FAST_SYNC mint
