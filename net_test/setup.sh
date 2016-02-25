#! /bin/bash

MACH="mempool"
NAME="net_test"
N=4

SEED=$(docker-machine ip ${MACH}1):46656
echo "seed node:" $SEED

# run the erisdb on each node
for i in `seq 1 $N`;
do
	dataI=$((i-1))
	machI=$i

	# lay the config with the seed for this session
	mintconfig --skip-upnp --seeds=$SEED > ./data/${NAME}_$dataI/config.toml

	eris --machine ${MACH}$machI chains new --dir $(pwd)/data/${NAME}_$dataI $NAME

	echo "%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%"
	echo " "
done   


# start the local proxy
mintconfig --skip-upnp --seeds=$SEED > ./data/local/config.toml
cp -r ./data/local ./data/local_data # so we don't contaminate the real dir
eris chains new --dir $(pwd)/data/local_data $NAME
