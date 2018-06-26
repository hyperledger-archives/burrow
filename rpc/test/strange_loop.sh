#!/usr/bin/env bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo -e "package test\n\nconst strangeLoopBytecodeHex = \"$(solc --bin "$script_dir/strange_loop.sol" | tail -1)\"" > "${script_dir}/strange_loop.go"
