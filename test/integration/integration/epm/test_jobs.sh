#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the test manager for monax jobs. It will run the testing
# sequence for monax jobs referencing test fixtures in this tests directory.

# ----------------------------------------------------------
# REQUIREMENTS

# monax installed locally

# ----------------------------------------------------------
# USAGE

# test_jobs.sh [appXX]


# Other variables
if [[ "$(uname -s)" == "Linux" ]]
then
  uuid=$(cat /proc/sys/kernel/random/uuid | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)
elif [[ "$(uname -s)" == "Darwin" ]]
then
  uuid=$(uuidgen | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1  | tr '[:upper:]' '[:lower:]')
else
  uuid="62d1486f0fe5"
fi

# Use the current built target, if it exists 
# Otherwise default to system wide executable
COMMIT_SHA=$(git rev-parse --short --verify HEAD)
cli_exec="$GOPATH/src/github.com/monax/monax/target/cli-${COMMIT_SHA}"
if ! [ -e $cli_exec ]
then
  cli_exec="monax"
fi

was_running=0
test_exit=0
chains_dir=$HOME/.monax/chains
name_base="monax-jobs-tests"
chain_name=$name_base-$uuid
name_full="$chain_name"_full_000
name_part="$chain_name"_participant_000
chain_dir=$chains_dir/$chain_name
repo=`DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"`


# ---------------------------------------------------------------------------
# Needed functionality

ensure_running(){
  if [[ "$($cli_exec ls --format {{.ShortName}} | grep $1)" == "$1" ]]
  then
    echo "$1 already started. Not starting."
    was_running=1
  else
    echo "Starting service: $1"
    $cli_exec services start $1 1>/dev/null
    early_exit
    sleep 3 # boot time
  fi
}

early_exit(){
  if [ $? -eq 0 ]
  then
    return 0
  fi

  echo "There was an error during setup; keys were not properly imported. Exiting."
  if [ "$was_running" -eq 0 ]
  then
    if [ "$ci" = true ]
    then
      $cli_exec services stop keys
    else
      $cli_exec services stop -rx keys
    fi
  fi
  exit 1
}

test_setup(){
  echo "Getting Setup"
  ensure_running keys

  # make a chain
  $cli_exec clean -y
  $cli_exec chains make --account-types=Full:1,Participant:1 $chain_name --unsafe
  key1_addr=$(cat $chain_dir/addresses.csv | grep $name_full | cut -d ',' -f 1)
  key2_addr=$(cat $chain_dir/addresses.csv | grep $name_part | cut -d ',' -f 1)
  key2_pub=$(cat $chain_dir/accounts.csv | grep $name_part | cut -d ',' -f 1)
  echo -e "Default Key =>\t\t\t\t$key1_addr"
  echo -e "Backup Key =>\t\t\t\t$key2_addr"
  $cli_exec chains start $chain_name --init-dir $chain_dir/$name_full 1>/dev/null
  sleep 5 # boot time
  chain_ip=$($cli_exec chains ip $chain_name)
  keys_ip=$($cli_exec services ip keys)
  echo -e "Chain at =>\t\t\t\t$chain_ip"
  echo -e "Keys at =>\t\t\t\t$keys_ip"
  echo "Setup complete"
}

goto_base(){
  cd $repo/jobs_fixtures
}

run_test(){
  # Run the jobs test
  echo ""
  echo -e "Testing $cli_exec jobs using fixture =>\t$1"
  goto_base
  cd $1
  if [ -z "$ci" ]
  then
    echo
    cat readme.md
    echo
    $cli_exec pkgs do --chain "$chain_name" --address "$key1_addr" --set "addr1=$key1_addr" --set "addr2=$key2_addr" --set "addr2_pub=$key2_pub" #--debug
  else
    echo
    cat readme.md
    echo
    $cli_exec pkgs do --chain "$chain_name" --address "$key1_addr" --set "addr1=$key1_addr" --set "addr2=$key2_addr" --set "addr2_pub=$key2_pub"
  fi
  test_exit=$?

  rm -rf ./abi &>/dev/null
  rm -rf ./bin &>/dev/null
  rm ./epm.output.json &>/dev/null
  rm ./jobs_output.csv &>/dev/null

  # Reset for next run
  goto_base
  return $test_exit
}

perform_tests(){
  echo ""
  goto_base
  apps=(app*/)
  for app in "${apps[@]}"
  do
    run_test $app

    # Set exit code properly
    test_exit=$?
    if [ $test_exit -ne 0 ]
    then
      failing_dir=`pwd`
      break
    fi
  done
}

perform_tests_that_should_fail(){
  echo ""
  goto_base
  apps=(expected-failure*/)
  for app in "${apps[@]}"
  do
    run_test $app

    # Set exit code properly
    test_exit=$?
    if [ $test_exit -ne 0 ]
    then
      # actually, this test is meant to pass
      test_exit=0
      break
    fi
  done
}

test_teardown(){
  if [ -z "$ci" ]
  then
    echo ""
    if [ "$was_running" -eq 0 ]
    then
      $cli_exec services stop -rx keys
    fi
    $cli_exec chains stop --force $chain_name 1>/dev/null
    # $cli_exec chains logs $chain_name -t 200 # uncomment me to dump recent VM/Chain logs
    # $cli_exec chains logs $chain_name -t all # uncomment me to dump all VM/Chain logs
    # $cli_exec chains logs $chain_name -t all | grep 'CALLDATALOAD\|Calling' # uncomment me to dump all VM/Chain logs and parse for Calls/Calldataload
    # $cli_exec chains logs $chain_name -t all | grep 'CALLDATALOAD\|Calling' > error.log # uncomment me to dump all VM/Chain logs and parse for Calls/Calldataload dump to a file
    $cli_exec chains rm --data $chain_name 1>/dev/null
    rm -rf $HOME/.monax/scratch/data/$name_base-*
    rm -rf $chain_dir
  else
    $cli_exec chains stop -f $chain_name 1>/dev/null
  fi
  echo ""
  if [ "$test_exit" -eq 0 ]
  then
    echo "Tests complete! Tests are Green. :)"
  else
    echo "Tests complete. Tests are Red. :("
  fi
  cd $start
  exit $test_exit
}

# ---------------------------------------------------------------------------
# Setup


echo "Hello! I'm the marmot that tests the $cli_exec jobs tooling."
echo
echo "testing with target $cli_exec"
echo
start=`pwd`
test_setup

# ---------------------------------------------------------------------------
# Go!

if [[ "$1" != "setup" ]]
then
  if ! [ -z "$1" ]
  then
    echo "Running One Test..."
    run_test "$1*/"
  else
    echo "Running tests that should fail"
    perform_tests_that_should_fail

    echo "Running tests that should pass"
    perform_tests
  fi
fi

# ---------------------------------------------------------------------------
# Cleaning up

if [[ "$1" != "setup" ]]
then
  test_teardown
fi
