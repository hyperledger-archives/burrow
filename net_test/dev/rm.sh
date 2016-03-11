#! /bin/bash
set -e

MACH="mempool"
NAME="net_test"
N=4

# run the erisdb on each node
for i in `seq 1 $N`;
do
	docker-machine ssh ${MACH}$i docker rm -vf \$\(docker ps -aq\)
	echo "%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%"
	echo " "
done   
