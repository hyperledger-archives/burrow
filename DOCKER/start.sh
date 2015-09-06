#! /bin/bash

ifExit(){
	if [ $? -ne 0 ]; then
		echo $1
		exit 1
	fi
}

export -f ifExit

#------------------------------------------------
# set and export directories

if [ "$CHAIN_ID" = "" ]; then
	echo "ecm requires CHAIN_ID be set"
	exit 1
fi

# TODO: deal with chain numbers
# and eg. $CONTAINER_NAME
CHAIN_DIR="/home/$USER/.eris/blockchains/$CHAIN_ID"

# set the tendermint directory
TMROOT=$CHAIN_DIR

if [ ! -d "$CHAIN_DIR" ]; then
	mkdir -p $CHAIN_DIR
	ifExit "Error making root dir $CHAIN_DIR"
fi

# our root chain
if [ ! $NODE_HOST ]; then
	NODE_HOST=interblock.io:46657
fi
if [ ! $NODE_ADDR ]; then
	NODE_ADDR=http://$NODE_HOST
fi

# where the etcb client scripts are
if [ ! $ECM_PATH ]; then
	ECM_PATH=.
fi

#------------------------------------------------
# dump key files if they are in env vars

if [ -z "$KEY" ]
then
  echo "No Key Given"
else
  echo "Key Given. Writing priv_validator.json"
	echo "$KEY" >> $CHAIN_DIR/priv_validator.json
fi

if [ -z "$GENESIS" ]
then
  echo "No Genesis Given"
else
  echo "Genesis Given. Writing genesis.json"
	echo "$GENESIS" > $CHAIN_DIR/genesis.json
fi

if [ -z "$GENESIS_CSV" ]
then
  echo "No Genesis_CSV Given"
else
  echo "Genesis_CSV Given. Writing genesis.csv"
  echo "$GENESIS_CSV" > $CHAIN_DIR/genesis.csv
fi

if [ -z "$CHAIN_CONFIG" ]
then
  echo "No Chain Config Given"
else
  echo "Chain Config Given. Writing config.toml"
	echo "$CHAIN_CONFIG" > $CHAIN_DIR/config.toml
fi

if [ -z "$SERVER_CONFIG" ]
then
  echo "No Server Config Given"
else
  echo "Server Config Given. Writing server_conf.toml"
	echo "$SERVER_CONFIG" > $CHAIN_DIR/server_conf.toml
fi

#------------------------------------------------
# export important vars

export TMROOT
export CHAIN_DIR
export NODE_ADDR
export NODE_HOST
export ECM_PATH  # set by Dockerfile

# print the version
bash $ECM_PATH/version.sh

#-----------------------------------------------------------------------
# either we are fetching a chain for the first time,
# creating one from scratch, or running one we already have
CMD=$1
case $CMD in
"install" ) $ECM_PATH/chain_install.sh
	;;
"new" ) $ECM_PATH/chain_new.sh
	;;
"run" ) $ECM_PATH/chain_run.sh
	;;
*)	echo "Enter a command for starting the chain (install, new, run, api)"
	;;
esac
