#! /bin/bash
set -e

MACH="mempool"
NAME="net_test"

TMROOT=data/local

# run a local erisdb node
erisdb $TMROOT  &> erisdb.log &

# set up eris-keys
# eris-keys server &
