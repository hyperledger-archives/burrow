#! /bin/bash

#-----------------------------------------------------------------------
# get genesis, seed, copy config

# get genesis if not already
if [ if -e "${CHAIN_DIR}/genesis.json" ]; then
	# etcb chain (given by $NODE_ADDR)
	REFS_CHAIN_ID=$(mintinfo --node-addr $NODE_ADDR genesis chain_id)
	ifExit "Error fetching default chain id from $NODE_ADDR"
	REFS_CHAIN_ID=$(echo "$REFS_CHAIN_ID" | tr -d '"') # remove surrounding quotes

	echo "etcb chain: $REFS_CHAIN_ID"

	# get the genesis.json for a refs chain from the /genesis rpc endpoint
	# for a different chain, use etcb (ie namereg on the ref chain)
	if [ "$CHAIN_ID" = "$REFS_CHAIN_ID" ] ; then
		# grab genesis.json and config
		mintinfo --node-addr $NODE_ADDR genesis > "${CHAIN_DIR}/genesis.json"
		ifExit "Error fetching genesis.json from $NODE_ADDR"
	else 
		# fetch genesis from etcb
		GENESIS=$(mintinfo --node-addr $NODE_ADDR names "${CHAIN_ID}_genesis.json" data)
		ifExit "Error fetching genesis.json for $CHAIN_ID: $GENESIS"

		echo $GENESIS > "${CHAIN_DIR}/genesis.json"

		SEED_NODE=$(mintinfo --node-addr $NODE_ADDR names "${CHAIN_ID}_seed" data)
		ifExit "Error grabbing seed node from $NODE_ADDR for $CHAIN_ID"
	fi
fi

# copy in config if not already
if [ ! -e "${CHAIN_DIR}/config.toml" ]; then
	cp $ECM_PATH/config.toml "${CHAIN_DIR}/config.toml"
	ifExit "Error copying config file from $ECM_PATH to $CHAIN_DIR"

	if [ "$SEED_NODE" = "" ]; then
		SEED_NODE=$NODE_ADDR
	fi

	if [ "$HOST_NAME" = "" ]; then
		HOST_NAME=mint_user
	fi
fi


# set seed node and host name
if [ "$SEED_NODE" != "" ]; then
	sed -i "s/^\(seeds\s*=\s*\).*\$/\1\"$SEED_NODE\"/" "${CHAIN_DIR}/config.toml"
	ifExit "Error setting seed node in config.toml"
	echo "Seed node: ${SEED_NODE}"
fi

if [ "$HOST_NAME" != "" ]; then
	sed -i "s/^\(moniker\s*=\s*\).*\$/\1\"$HOST_NAME\"/" "${CHAIN_DIR}/config.toml"
	ifExit "Error setting host name in config.toml"
fi

#-----------------------------------------------------------------------

# would be nice if we could stop syncing once we're caught up ...
echo "Running mint in ${CHAIN_DIR}"
tendermint node --fast_sync 
ifExit "Error running tendermint!"
