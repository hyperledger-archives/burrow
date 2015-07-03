#! /bin/sh

echo "new chain: $CHAIN_ID"

if [ "$GENERATE_GENESIS" = "true" ]; then
	mintgen --single $CHAIN_DIR
fi

if [ "$RUN" = "true" ]; then
	tendermint node	
else
	# this will just run for a second and quit
	tendermint node & last_pid=$! && sleep 1 && kill -KILL $last_pid
fi
