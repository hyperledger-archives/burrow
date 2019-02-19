#!/bin/bash

# This script provides a simple running test chain that will generate height notification events
# Run postgres in background with: docker run -p 5432:5432 postgres

vent_test_dir="$GOPATH/src/github.com/hyperledger/burrow/vent/test"
[[ ! -f burrow.toml ]] && burrow spec -f1 | burrow configure -s- > burrow.toml && rm -rf .burrow
burrow start -v0 &> burrow.log &
sleep 2s
burrow vent --db-block --abi "$vent_test_dir/EventsTest.abi" --spec "$vent_test_dir/sqlsol_example.json"
# Now:
# psql -h 127.0.0.1 -p 5432 -U postgres
# LISTEN height;
# -- run any command to see notifications:
# SELECT true;
# -- run it some more

# Generate some other events (on channels meta and keyed_meta) with:
# addr=$(curl -s localhost:26658/validators | jq -r '.result.BondedValidators[0].Address')
# burrow deploy -a $addr

trap "killall burrow" EXIT
