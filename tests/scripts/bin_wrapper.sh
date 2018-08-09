#!/usr/bin/env bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Use burrow, solc binaries in the repo's bin directory
export PATH=$(readlink -f ${script_dir}/../../bin):$PATH

"$@"
