#!/usr/bin/env bash

chain=$(mktemp -d)
cd $chain
$burrow_bin spec -v1 -d2 | $burrow_bin configure -s- --curve-type secp256k1 > burrow.toml
$burrow_bin start &> /dev/null &
burrow_pid=$!

contracts=$(mktemp -d)
cd $contracts

function finish {
    kill -TERM $burrow_pid
    rm -rf "$chain"
    rm -rf "$contracts"
}
trap finish EXIT

npm install -g truffle
truffle unbox metacoin

cat << EOF > truffle-config.js
module.exports = {
  networks: {
   burrow: {
     host: "127.0.0.1",
     port: 26660,
     network_id: "*",
   },
  }
};
EOF
truffle test --network burrow
