#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the test runner for burrow integration tests.
# It is responsible for starting up a single node burrow test chain and tearing it down afterwards.

# ----------------------------------------------------------
# REQUIREMENTS

# * GNU parallel
# * jq

# ----------------------------------------------------------
# USAGE
# source test_runner.sh

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export burrow_bin=${burrow_bin:-burrow}
export solc_bin=${solc_bin:-solc}
export solang_bin=${solang_bin:-solang}

# If false we will not try to start Burrow and expect them to be running
export boot=${boot:-true}
export debug=${debug:-false}
export clean=${clean:-true}

export test_exit=0

if [[ "$debug" = true ]]; then
    set -o xtrace
fi


# Note: do not set -e in order to capture exit correctly in mocha
# ----------------------------------------------------------
# Constants

# Ports etc must match those in burrow.toml
export BURROW_HOST=127.0.0.1
export BURROW_GRPC_PORT=20123


export chain_dir="$script_dir/chain"
export burrow_root="$chain_dir/.burrow"

# Temporary logs
export burrow_log="$chain_dir/burrow.log"
#
# ----------------------------------------------------------

# ---------------------------------------------------------------------------
# Needed functionality

pubkey_of() {
    jq -r ".Accounts | map(select(.Name == \"$1\"))[0].PublicKey.PublicKey" "$chain_dir/genesis.json"
}

address_of() {
    jq -r ".Accounts | map(select(.Name == \"$1\"))[0].Address" "$chain_dir/genesis.json"
}

test_setup(){
  echo "Setting up..."
  cd "$script_dir"

  echo
  echo "Using binaries:"
  echo "  $(type ${solc_bin}) (version: $(${solc_bin} --version))"
  echo "  $(type ${solang_bin}) (version: $(${solang_bin} --version))"
  echo "  $(type ${burrow_bin}) (version: $(${burrow_bin} --version))"
  echo
  # start test chain
  BURROW_ADDRESS="$BURROW_HOST:$BURROW_GRPC_PORT"
  if [[ "$boot" = true ]]; then
    echo "Starting Burrow using GRPC address: $BURROW_ADDRESS..."
    echo
    rm -rf ${burrow_root}
    pushd "$chain_dir"
    ${burrow_bin} start --index 0 --grpc-address $BURROW_ADDRESS 2> "$burrow_log"&
    burrow_pid=$!
    popd
  else
    echo "Not booting Burrow, but expecting Burrow to be running with tm RPC on port $BURROW_GRPC_PORT"
  fi

  export key1_addr=$(address_of "Full_0")
  export key2_addr=$(address_of "Participant_0")
  export key1=Full_0
  export key2=Participant_0
  export key2_pub=$(pubkey_of "Participant_0")

  echo -e "Default Key =>\t\t\t\t$key1_addr"
  echo -e "Backup Key =>\t\t\t\t$key2_addr"
  sleep 4 # boot time

  echo "Setup complete"
  echo ""
}

test_teardown(){
  echo "Cleaning up..."
  if [[ "$boot" = true ]]; then
    echo "Killing burrow with PID $burrow_pid"
    kill ${burrow_pid} 2> /dev/null
    echo "Waiting for burrow to shutdown..."
    wait ${burrow_pid} 2> /dev/null
    rm -rf "$burrow_root"
  fi
  echo ""
  if [[ "$test_exit" -eq 0 ]]
  then
    [[ "$boot" = true ]] && rm -f "$burrow_log"
    echo "Tests complete! Tests are Green. :)"
  else
    echo "Tests complete. Tests are Red. :("
   fi
  exit ${test_exit}
}

