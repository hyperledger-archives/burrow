#! /bin/sh

# if no CHAIN_ID, get that off the main test net
if [[ ! $CHAIN_ID ]]; then
	# get the chain id 
	CHAIN_ID=$(mintinfo --node-addr $NODE_ADDR genesis chain_id)
	ifExit "Error fetching default chain id from $NODE_ADDR"
	CHAIN_ID=$(echo "$CHAIN_ID" | tr -d '"') # remove surrounding quotes
fi

CHAIN_DIR="${ROOT_DIR}/$CHAIN_ID"

if [[ ! -d  $CHAIN_DIR ]]; then
	echo "Unknown chain ($CHAIN_ID)"
	exit 1
fi

echo Running chain $CHAIN_ID
tendermint node
