#! /bin/bash

echo "your new chain, kind marmot: $CHAIN_ID"

# lay the genesis
# if it exists, just overwrite the chain id
if [ ! -f $CHAIN_DIR/genesis.json ]; then
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


# if no config was given, lay one with the given options
if [ ! -f $CHAIN_DIR/config.toml ]; then
	echo "running mintconfig $CONFIG_OPTS"
	mintconfig $CONFIG_OPTS > $CHAIN_DIR/config.toml
else
	echo "found config file:"
	cat $CHAIN_DIR/config.toml
fi

# if an address is given, keys service should have the priv key
if [ "$REGISTER_ADDRESS" != "" ]; then
	echo "registering $CHAIN_ID with the etcb_testnet at interblock.io from address $REGISTER_ADDRESS"

	# register the genesis
	mintx --node-addr http://interblock.io:46657/ --sign-addr http://keys:4767 --addr $REGISTER_ADDRESS name --name "$CHAIN_ID:genesis" --data $(cat $CHAIN_DIR/genesis.json) --amt 10000 --fee 0 --sign --broadcast --wait
	ifExit "Error registering genesis with etcb_testnet"

	# register the seed/s
	mintx --node-addr http://interblock.io:46657/ --sign-addr http://keys:4767 --addr $REGISTER_ADDRESS name --name "$CHAIN_ID:seeds" --data $NEW_P2P_SEEDS --amt 10000 --fee 0 --sign --broadcast --wait
	ifExit "Error registering seeds with etcb_testnet"
fi

# run the node.
# TODO: maybe bring back this stopping option if we think its useful
# tendermint node & last_pid=$! && sleep 1 && kill -KILL $last_pid
if [ $ERISDB_API ]; then
	echo "Running chain $CHAIN_ID (via ErisDB API)"
	erisdb $TMROOT
	ifExit "Error starting erisdb"
else
	echo Running chain $CHAIN_ID
	tendermint node
	ifExit "Error starting tendermint"
fi
