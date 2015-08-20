#! /bin/bash

echo "new chain: $CHAIN_ID"

if [ "$GENERATE_GENESIS" = "true" ]; then
	mintgen random --dir="/home/eris/.eris/blockchains/$CHAIN_ID" 1 $CHAIN_ID
	ifExit "Error creating genesis file"
fi

if [ "$RUN" = "true" ]; then
	tendermint node
	ifExit "Error starting tendermint"
else
	# this will just run for a second and quit
	tendermint node & last_pid=$! && sleep 1 && kill -KILL $last_pid
	ifExit "Error starting tendermint"
fi
