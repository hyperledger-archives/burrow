#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the test manager for monax jobs. It will run the testing
# sequence for monax jobs referencing test fixtures in this tests directory.

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

run_test(){
  # Run the jobs test
  (
  echo ""
  echo -e "Testing $burrow_bin jobs using fixture =>\t$1"
  goto_base

  cd $1
  echo "PWD: $PWD"
  cat readme.md
  echo

#
  deploy_cmd="${burrow_bin} deploy --chain-url=$BURROW_HOST:$BURROW_GRPC_PORT --keys=$BURROW_HOST:$BURROW_GRPC_PORT \
   --address $key1_addr --set addr1=$key1_addr --mempool-signing --set addr2=$key2_addr --set addr2_pub=$key2_pub"
  [[ "$debug" == true ]] && deploy_cmd="$deploy_cmd --debug"
  echo "executing deploy with command line:"
  echo "$deploy_cmd"
  eval "${deploy_cmd}"
  )
}

perform_tests(){
  echo ""
  goto_base
  apps=($1*/)
  # Run all jobs in parallel with mempool signing
  parallel --joblog "$job_log" -k -j 100 run_test ::: "${apps[@]}" 2>&1 | tee "$test_output"
  failures=($(awk 'NR==1{for(i=1;i<=NF;i++){if($i=="Exitval"){c=i;break}}} ($c=="1"&& NR>1){print $NF}' "$job_log"))
  test_exit=${#failures[@]}
}

perform_tests_that_should_fail(){
  echo ""
  goto_base
  apps=($1*/)
  perform_tests "$1"
  expectedFailures="${#apps[@]}"
  if [[ "$test_exit" -eq $expectedFailures ]]
  then
    echo "Success! We go the correct number of failures: $test_exit (don't worry about messages above)"
    echo
    test_exit=0
  else
    echo "Expected $expectedFailures but only got $test_exit failures"
    test_exit=$(expr ${expectedFailures} - ${test_exit})
  fi
}

export -f run_test goto_base

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
