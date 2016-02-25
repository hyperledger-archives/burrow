#! /bin/bash


MACH="mempool"
NAME="net_test"
N=4

SEED=$(docker-machine ip ${MACH}1):46656
echo "seed node:" $SEED

# copy data to remotes
for i in `seq 0 $((N-1))`;
do

	dataI=$i
	machI=$((i+1))

	# lay the config with the seed for this session
	mintconfig --skip-upnp --seeds=$SEED > ./data/${NAME}_$dataI/config.toml

	# copy the startup bash script into the data ...
	cp ./data/init.sh ./data/${NAME}_$dataI/init.sh

	# clear and copy the node data
	docker-machine ssh ${MACH}$machI rm -rf $NAME
	docker-machine scp -r ./data/${NAME}_$dataI ${MACH}$machI:$NAME

	# docker-machine ssh ${MACH}$machI ls $NAME
	echo "###############"
done   


