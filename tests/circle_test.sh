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
# Run integration tests

tests/test_client.sh
test_exit=$?

# ----------------------------------------------------------
# Cleanup

echo
echo
echo "Cleaning up"
cd $start
exit $test_exit
