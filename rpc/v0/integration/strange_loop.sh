#!/usr/bin/env bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo -e "package integration\n\nconst strangeLoopBytecode = \"$(solc --bin rpc/v0/integration/strange_loop.sol | tail -1)\"" > "${script_dir}/strange_loop.go"
