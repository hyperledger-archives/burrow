#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the test manager for playbooks. It will run the testing
# sequence for playbooks referencing test fixtures in this tests directory.

# ----------------------------------------------------------
# REQUIREMENTS

#

# ----------------------------------------------------------
# USAGE

# deploy.sh [appXX]

export script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# For parallel when default shell is not bash (we need exported functions)
export SHELL=$(which bash)
export job_log="$script_dir/deploy-test-log.txt"
export test_output="$script_dir/deploy-test-output.txt"

source "$script_dir/test_runner.sh"

goto_base(){
  cd ${script_dir}/jobs_fixtures
}

perform_tests(){
  echo ""
  goto_base
  apps=$1*/deploy.yaml

  # Run all jobs in parallel with proxy signing
  deploy_cmd="${burrow_bin} deploy --jobs=100 --chain=$BURROW_HOST:$BURROW_GRPC_PORT --keys-dir="${script_dir}/keys" \
   --address $key1 --proxy-signing --set key2_addr=$key2_addr --set addr2_pub=$key2_pub --set key1=$key1 --set key2=$key2 --proposal-create $apps"
  [[ "$debug" == true ]] && deploy_cmd="$deploy_cmd --debug"
  echo "executing deploy with command line:"
  echo "$deploy_cmd"
  ${deploy_cmd}
  test_exit=$?
}

perform_tests_that_should_fail(){
  echo ""
  goto_base
  apps=($1*/)
  perform_tests "$1"
  expectedFailures="${#apps[@]}"
  if [[ "$test_exit" -eq $expectedFailures ]]
  then
    echo "Success! We go the correct number of failures: ${test_exit} (don't worry about messages above)"
    echo
    test_exit=0
  else
    echo "Expected $expectedFailures but only got $test_exit failures"
    test_exit=$(expr ${expectedFailures} - ${test_exit})
  fi
}

export -f goto_base

deploy_tests(){
  echo "Hello! I'm the marmot that tests the $burrow_bin tooling."
  echo
  echo "testing with target $burrow_bin"
  echo
  test_setup
  # Cleanup
  cleanup() {
    goto_base

    if [[ "$clean" == true ]]
    then
      git clean -fdxq "${script_dir}/jobs_fixtures" "${script_dir}/keys"
      if [[ "$test_exit" -eq 0 ]]
      then
          rm -f "$job_log" "$test_output"
      fi
    fi
    # This exits so must be last thing called
    test_teardown
  }
  trap cleanup EXIT
  if ! [ -z "$1" ]
  then
    echo "Running tests beginning with $1..."
    perform_tests "$1"
  else
    echo "Running tests that should fail"
    perform_tests_that_should_fail expected-failure

    if [[ "$test_exit" -eq 0 ]]
    then
      echo "Running tests that should pass"
      perform_tests app
    fi
  fi
}

deploy_tests "$1"
