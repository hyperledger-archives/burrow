#! /bin/bash

echo "your new chain, kind marmot: $CHAIN_ID"

if [ "$GENERATE_GENESIS" = "true" ]; then
	if [ "$CSV" = "" ]; then
		mintgen random --dir="$CHAIN_DIR" 1 $CHAIN_ID
		ifExit "Error creating random genesis file"
	else
		mintgen known --csv="$CSV" $CHAIN_ID > $CHAIN_DIR/genesis.json
		ifExit "Error creating genesis file from csv"
	fi
else
	# apparently just outputing to $CHAIN_DIR/genesis.json doesn't work so we copy
	cat $CHAIN_DIR/genesis.json | jq .chain_id=\"$CHAIN_ID\" > genesis.json
	cp genesis.json $CHAIN_DIR/genesis.json
fi

mintconfig $CONFIG_OPTS > $CHAIN_DIR/config.toml

if [ "$RUN" = "true" ]; then
	tendermint node
	ifExit "Error starting tendermint"
else
	# this will just run for a second and quit
	tendermint node & last_pid=$! && sleep 1 && kill -KILL $last_pid
	ifExit "Error starting tendermint"
fi
