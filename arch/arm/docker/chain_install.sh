#! /bin/bash

#-----------------------------------------------------------------------
# get genesis, seed, copy config

export MINTX_NODE_ADDR=$NODE_ADDR

# get genesis if not already
if [ ! -e "${CHAIN_DIR}/genesis.json" ]; then
	# etcb chain (given by $NODE_ADDR)
	REFS_CHAIN_ID=$(mintinfo genesis chain_id)
	ifExit "Error fetching default chain id from $NODE_ADDR"
	REFS_CHAIN_ID=$(echo "$REFS_CHAIN_ID" | tr -d '"') # remove surrounding quotes

	echo "etcb chain: $REFS_CHAIN_ID"

	# get the genesis.json for a refs chain from the /genesis rpc endpoint
	# for a different chain, use etcb (ie namereg on the ref chain)
	if [ "$CHAIN_ID" = "$REFS_CHAIN_ID" ] ; then
		# grab genesis.json 
		mintinfo genesis > "${CHAIN_DIR}/genesis.json"
		ifExit "Error fetching genesis.json from $NODE_ADDR"
	else 
		# fetch genesis from etcb
		GENESIS=$(mintinfo names "${CHAIN_ID}/genesis" data)
		ifExit "Error fetching genesis.json for $CHAIN_ID: $GENESIS"

		echo $GENESIS > "${CHAIN_DIR}/genesis.json"

		SEED_NODE=$(mintinfo names "${CHAIN_ID}/seeds" data)
		ifExit "Error grabbing seed node from $NODE_ADDR for $CHAIN_ID"
	fi
fi

# copy in config if not already
if [ ! -e "${CHAIN_DIR}/config.toml" ]; then
	echo "laying default config..."
	mintconfig > $CHAIN_DIR/config.toml
	ifExit "Error creating config"

	if [ "$SEED_NODE" = "" ]; then
		SEED_NODE=$P2P_ADDR
	fi

	if [ "$HOST_NAME" = "" ]; then
		HOST_NAME=mint_user
	fi
fi

# set seed node and host name
if [ "$SEED_NODE" != "" ]; then
	echo "Seed node: $SEED_NODE"
	# NOTE the NODE_ADDR must not have any slashes (no http://)
	sed -i "s/^\(seeds\s*=\s*\).*\$/\1\"$SEED_NODE\"/" "${CHAIN_DIR}/config.toml"
	ifExit "Error setting seed node in config.toml"
fi

if [ "$HOST_NAME" != "" ]; then
	echo "Host name: $HOST_NAME"
	sed -i "s/^\(moniker\s*=\s*\).*\$/\1\"$HOST_NAME\"/" "${CHAIN_DIR}/config.toml"
	ifExit "Error setting host name in config.toml"
fi

#-----------------------------------------------------------------------

# would be nice if we could stop syncing once we're caught up ...
echo "Running mint in ${CHAIN_DIR}"
tendermint node --fast_sync 
ifExit "Error running tendermint!"
