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
if [ ! $NODE_ADDR ]; then
	NODE_ADDR=http://interblock.io:46657
fi

# where the etcb client scripts are
if [ ! $ECM_PATH ]; then
	ECM_PATH=.
fi


export TMROOT
export CHAIN_DIR
export NODE_ADDR
export ECM_PATH  # set by Dockerfile


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
*)	echo "Enter a command for starting the chain (install, new, run)"
	;;
esac
