#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the test manager for monax jobs. It will run the testing
# sequence for monax jobs referencing test fixtures in this tests directory.

# ----------------------------------------------------------
# REQUIREMENTS

# m

# ----------------------------------------------------------
# USAGE

# run_pkgs_tests.sh [appXX]


export script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "$script_dir/test_runner.sh"

export js_dir="${script_dir}/../burrow.js"


perform_js_tests(){
  cd "$js_dir"
  test_account="{\"address\": \"$key1_addr\"}"
  echo "Using test account:"
  echo "$test_account"
  account="$test_account" mocha --bail --exit --recursive ${1}
  test_exit=$?
}

burrowjs_tests() {
    echo "Hello! I'm the marmot that tests burrow-js."
    echo

    test_setup
    trap test_teardown EXIT

    echo "Running burrow.js tests..."
    perform_js_tests "$1"
}

burrowjs_tests "$1"
