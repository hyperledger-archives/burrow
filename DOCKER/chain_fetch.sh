#! /bin/sh

ifExit(){
	if [ $? -ne 0 ]; then
		echo $1
		exit 1
	fi
}

# fetching a chain means grabbing its genesis.json (from etcb) and putting it in the right folder
# we also copy in the config.toml, and update grab the seed node from etcb

## get chain id for main reference chain (given by NODE_ADDR)
REFS_CHAIN_ID=$(mintinfo --node-addr $NODE_ADDR genesis chain_id)
ifExit "Error fetching default chain id from $NODE_ADDR"
REFS_CHAIN_ID=$(echo "$REFS_CHAIN_ID" | tr -d '"') # remove surrounding quotes

# get the genesis.json for a refs chain from the /genesis rpc endpoint
# for a different chain, use etcb (ie namereg on the ref chain)
if [ "$CHAIN_ID" = "$REF_CHAIN_ID"    ] ; then
	# grab genesis.json and config
	mintinfo --node-addr $NODE_ADDR genesis > "${CHAIN_DIR}/genesis.json"
	ifExit "Error fetching genesis.json from $NODE_ADDR"
	cp $ECM_PATH/config.toml "${CHAIN_DIR}/config.toml"
	ifExit "Error copying config file from $ECM_PATH to $CHAIN_DIR"
else 
	# fetch genesis from etcb
	GENESIS=$(mintinfo --node-addr $NODE_ADDR names "${CHAIN_ID}_genesis.json" data)
	ifExit "Error fetching genesis.json for $CHAIN_ID: $GENESIS"

	echo $GENESIS > "${CHAIN_DIR}/genesis.json"
	cp $ECM_PATH/config.toml "${CHAIN_DIR}/config.toml"
	ifExit "Error copying config file from $ECM_PATH to $CHAIN_DIR"

	SEED_NODE=$(mintinfo --node-addr $NODE_ADDR names "${CHAIN_ID}_seed" data)
	ifExit "Error grabbing seed node from $NODE_ADDR for $CHAIN_ID"

	sed -i "s/^\(seeds\s*=\s*\).*\$/\1\"$SEED_NODE\"/" "${CHAIN_DIR}/config.toml"
	ifExit "Error setting seed node in config.toml"
	echo "Seed node: ${SEED_NODE}"
fi

if [ "$HOST_NAME" = "" ]; then
	HOST_NAME=mint_user
fi

echo "Host name: ${HOST_NAME}"

sed -i "s/^\(moniker\s*=\s*\).*\$/\1\"$HOST_NAME\"/" "${CHAIN_DIR}/config.toml"
ifExit "Error setting host name in config.toml"


# would be nice if we could stop syncing once we're caught up ...
echo "Running mint in ${CHAIN_DIR}"
tendermint node --fast_sync 
ifExit "Error running tendermint!"
