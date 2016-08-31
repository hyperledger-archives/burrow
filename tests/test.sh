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
  echo "Building eris-cm in a docker container."
  set -e
  tests/build_tool.sh 1>/dev/null
  set +e
  if [ $? -ne 0 ]
  then
    echo "Could not build eris-cm. Debug via by directly running [`pwd`/tests/build_tool.sh]"
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

check_test(){
  # check chain is running
  chain=( $(eris chains ls --quiet --running | grep $uuid) )
  if [ ${#chain[@]} -ne 1 ]
  then
    echo "chain does not appear to be running"
    echo
    ls -la $dir_to_use
    test_exit=1
    return 1
  fi

  # check results file exists
  if [ ! -e "$chains_dir/$uuid/accounts.csv" ]
  then
    echo "accounts.csv not present"
    ls -la $chains_dir/$uuid
    pwd
    ls -la $chains_dir
    test_exit=1
    return 1
  fi

  # check genesis.json
  genOut=$(cat $dir_to_use/genesis.json | sed 's/[[:space:]]//g')
  genIn=$(eris chains plop $uuid genesis | sed 's/[[:space:]]//g')
  if [[ "$genOut" != "$genIn" ]]
  then
    test_exit=1
    echo "genesis.json's do not match"
    echo
    echo "expected"
    echo
    echo -e "$genOut"
    echo
    echo "received"
    echo
    echo -e "$genIn"
    echo
    echo "difference"
    echo
    diff  <(echo "$genOut" ) <(echo "$genIn") | colordiff
    return 1
  fi

  # check priv_validator
  privOut=$(cat $dir_to_use/priv_validator.json | tr '\n' ' ' | sed 's/[[:space:]]//g' | set 's/(,\"last_height\":[^0-9]+,\"last_round\":[^0-9]+,\"last_step\":[^0-9]+//g' )
  privIn=$(eris data exec $uuid "cat /home/eris/.eris/chains/$uuid/priv_validator.json" | tr '\n' ' ' | sed 's/[[:space:]]//g' | set 's/(,\"last_height\":[^0-9]+,\"last_round\":[^0-9]+,\"last_step\":[^0-9]+//g' )
  if [[ "$privOut" != "$privIn" ]]
  then
    test_exit=1
    echo "priv_validator.json's do not match"
    echo
    echo "expected"
    echo
    echo -e "$privOut"
    echo
    echo "received"
    echo
    echo -e "$privIn"
    echo
    echo "difference"
    echo
    diff  <(echo "$privOut" ) <(echo "$privIn") | colordiff
    return 1
  fi
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
# perform_client_tests

# ---------------------------------------------------------------------------
# Cleaning up

test_teardown