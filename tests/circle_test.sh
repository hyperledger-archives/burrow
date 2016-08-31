#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the test manager for epm to be ran from circle ci.
# It will run the testing sequence for eris-db using docker.

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally
# docker-machine installed locally
# eris installed locally
# jq installed locally

# ----------------------------------------------------------
# USAGE

# circle_test.sh

# ----------------------------------------------------------
# Set defaults

uuid=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)
machine="eris-test-edb-$uuid"
ver=$(cat version/version.go | tail -n 1 | cut -d ' ' -f 4 | tr -d '"')
start=`pwd`

# ----------------------------------------------------------
# Get machine sorted

echo "Setting up a Machine for eris-cm Testing"
docker-machine create --driver amazonec2 $machine 1>/dev/null
if [ "$?" -ne 0 ]
then
  echo "Failed to create The Machine for eris-db Testing"
  exit 1
fi
docker-machine scp tests/docker.sh ${machine}:
if [ "$?" -ne 0 ]
then
  echo "Failed to copy the 'docker.sh' script into the container"
  exit 1
fi
docker-machine ssh $machine sudo env DOCKER_VERSION=$DOCKER_VERSION '$HOME/docker.sh'
if [ "$?" -ne 0 ]
then
  echo "Failed to install Docker client into the container"
  exit 1
fi
eval $(docker-machine env $machine)
echo "Machine setup."
echo
docker version
echo

# ----------------------------------------------------------
# Run integration tests

tests/test_client.sh
test_exit=$?

# ----------------------------------------------------------
# Cleanup

echo
echo
echo "Cleaning up"
docker-machine rm --force $machine
cd $start
exit $test_exit
