#! /bin/bash
set -e

MACH="mempool"
NAME="net_test"
N=4

# run the erisdb on each node
for i in `seq 1 $N`;
do
	dataI=$((i-1))
	machI=$i

	docker-machine ssh ${MACH}$machI docker run --name erisdb -d -p 46656-46657:46656-46657 -p 1337:1337 -v \$\(pwd\)/net_test:/home/eris/data quay.io/eris/erisdb-dev &
	echo "%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%"
	echo " "
done   
