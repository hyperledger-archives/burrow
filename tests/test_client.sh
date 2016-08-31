#!/usr/bin/env bash

# Copyright 2015, 2016 Eris Industries (UK) Ltd.
# This file is part of Eris-RT

# Eris-RT is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.

# Eris-RT is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

# ----------------------------------------------------------
# PURPOSE

# This is the integration test manager for eris-db. It will
# run the integration testing sequence for eris-db using docker
# and the dependent eris components within the eris platform
# for eris-db.  Specifically eris-db and the eris-db client
# require a key management component for signing transactions
# and validating blocks.

# ----------------------------------------------------------
# REQUIREMENTS

# eris installed locally

# ----------------------------------------------------------
# USAGE

# test.sh

# ----------------------------------------------------------
# Set defaults

# Where are the Things?

name=eris-db
base=github.com/eris-ltd/$name
repo=`pwd`
if [ "$CIRCLE_BRANCH" ]
then
  ci=true
  linux=true
elif [ "$TRAVIS_BRANCH" ]
then
  ci=true
  osx=true
elif [ "$APPVEYOR_REPO_BRANCH" ]
then
  ci=true
  win=true
else
  repo=$GOPATH/src/$base
  ci=false
fi

branch=${CIRCLE_BRANCH:=master}
branch=${branch/-/_}
branch=${branch/\//_}

# Other variables
was_running=0
test_exit=0
chains_dir=$HOME/.eris/chains

export ERIS_PULL_APPROVE="true"
export ERIS_MIGRATE_APPROVE="true"

# ---------------------------------------------------------------------------
# Needed functionality

ensure_running(){
  if [[ "$(eris services ls -qr | grep $1)" == "$1" ]]
  then
    echo "$1 already started. Not starting."
    was_running=1
  else
    echo "Starting service: $1"
    eris services start $1 1>/dev/null
    early_exit
    sleep 3 # boot time
  fi
}

early_exit(){
  if [ $? -eq 0 ]
  then
    return 0
  fi

  echo "There was an error duing setup; keys were not properly imported. Exiting."
  if [ "$was_running" -eq 0 ]
  then
    if [ "$ci" = true ]
    then
      eris services stop keys
    else
      eris services stop -r keys
    fi
  fi
  exit 1
}

get_uuid() {
  if [[ "$(uname -s)" == "Linux" ]]
  then
    uuid=$(cat /proc/sys/kernel/random/uuid | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)
  elif [[ "$(uname -s)" == "Darwin" ]]
  then
    uuid=$(uuidgen | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)
  else
    uuid="2231587f0fe5"
  fi
  echo $uuid
}

test_build() {
  echo ""
  echo "Building eris-db in a docker container."
  set -e
  tests/build_tool.sh 1>/dev/null
  set +e
  if [ $? -ne 0 ]
  then
    echo "Could not build eris-db. Debug via by directly running [`pwd`/tests/build_tool.sh]"
    exit 1
  fi
  echo "Build complete."
  echo ""
}

test_setup(){
  echo "Getting Setup"
  if [ "$ci" = true ]
  then
    eris init --yes --pull-images=true --testing=true 1>/dev/null
  fi

  ensure_running keys
  echo "Setup complete"
}

start_chain(){
  echo
  echo "starting new chain for client tests..."
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi
  eris chains make $uuid --account-types=Participant:2,Validator:1
  eris chains new $uuid --dir "$uuid"/"$uuid"_validator_000
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi
  sleep 3 # let 'er boot

  # set variables for chain
  CHAIN_ID=$uuid
  eris_client_ip=$(eris chains inspect $uuid NetworkSettings.IPAddress)
  ERIS_CLIENT_NODE_ADDRESS="tcp://$(eris chains inspect $uuid NetworkSettings.IPAddress):46657"
  ERIS_CLIENT_SIGN_ADDRESS="http://$(eris services inspect keys NetworkSettings.IPAddress):4767"
  echo "node address: " $ERIS_CLIENT_NODE_ADDRESS
  echo "keys address: " $ERIS_CLIENT_SIGN_ADDRESS

  # set addresses from participants
  participant_000_address=$(cat $chains_dir/accounts.json | jq '. | ."$uuid"_participant_000.address')
  participant_001_address=$(cat $chains_dir/accounts.json | jq '. | ."$uuid"_participant_001.address')
 }

 stop_chain(){
  echo
  echo "stopping test chain for client tests..."
  eris chains stop --force $uuid
  if [ ! "$ci" = true ]
  then
    eris chains rm --data $uuid
  fi
  rm -rf $HOME/.eris/scratch/data/$uuid
  rm -rf $chains_dir/$uuid
}

perform_client_tests(){
  uuid=$(get_uuid)
  start_chain

  echo
  echo "simplest client send transaction test"
  amount=1000
  eris-client tx send --amt $amount -addr $participant_000_address --to $participant_001_address
  sleep 2 # poll for resulting state - sleeping, rather than waiting for confirmation
  sender_amt=$(curl "$eris_client_ip"/get_account?address=$participant_000_address | jq '. | .result[1].account.balance')
  receiver_amt=$(curl "$eris_client_ip"/get_account?address=$participant_001_address | jq '. | .result[1].account.balance')
  difference='expr $receiver_amt - $sender_amt'
  if [[ "$difference" != "$amount" ]]
  then
  	echo "simple send transaction failed"
    return 1
  fi
  echo
  stop_chain
}

test_teardown(){
  if [ "$ci" = false ]
  then
    echo
    if [ "$was_running" -eq 0 ]
    then
      eris services stop -rx keys
    fi
    echo
  fi
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
# Get the things build and dependencies turned on

echo "Hello! I'm the marmot that tests the eris-db tooling"
start=`pwd`
cd $repo
test_setup
test_build
echo

# ---------------------------------------------------------------------------
# Go ahead with node integration tests 

# TODO

# ---------------------------------------------------------------------------
# Go ahead with client integration tests !

echo "Running Client Tests..."
perform_client_tests

# ---------------------------------------------------------------------------
# Cleaning up

test_teardown