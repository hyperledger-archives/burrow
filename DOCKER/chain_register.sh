#! /bin/bash

echo "registering $CHAIN_ID in the grand registry of marmot mayhem and marmalade"

# lay the genesis
# if it exists, just overwrite the chain id
if [ ! -f $CHAIN_DIR/genesis.json ]; then
	"Could not find genesis file in $CHAIN_DIR. Did you run `eris chains new $CHAIN_ID`?"
	exit 1
fi


echo "or less dramatically, registering $CHAIN_ID with the $ETCB_CHAIN_ID chain at $MINTX_NODE_ADDR from address $PUBKEY"

echo "NAME ${CHAIN_ID}_genesis"
cat $CHAIN_DIR/genesis.json

# register the genesis
RES=`mintx name --pubkey=$PUBKEY --name="${CHAIN_ID}/genesis" --data-file=$CHAIN_DIR/genesis.json --amt=10000 --fee=0 --sign --broadcast --chainID=$ETCB_CHAIN_ID --wait`
ifExit "$RES" "Error registering genesis with etcb_testnet"
echo $RES | grep "Incorrect"
if0Exit "$RES" "Error registering genesis with etcb_testnet"
echo $RES

# register the seed/s
RES=`mintx name --pubkey=$PUBKEY --name="${CHAIN_ID}/seeds" --data="$NEW_P2P_SEEDS" --amt=10000 --fee=0 --sign --broadcast --chainID=$ETCB_CHAIN_ID --wait`
ifExit "$RES" "Error registering seeds with etcb_testnet"
echo $RES | grep "Incorrect"
if0Exit "$RES" "Error registering seeds with etcb_testnet"
echo $RES
