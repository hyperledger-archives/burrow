#!/usr/bin/env bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
solidity_file="$1"
solidity_name=$(basename $1)
go_file="$2"
package=$(basename $(dirname "$go_file"))
solidity_bin=$(solc --bin "$solidity_file" | tail -1)

cat << GOFILE > "$go_file"
package ${package}

import "github.com/tmthrgd/go-hex"

var Bytecode_${solidity_name%%.sol} = hex.MustDecodeString("${solidity_bin}")
GOFILE

