#!/bin/bash

echo "Cleaning up old test results"

rm -rf orphan-test-db 2>/dev/null
rm merkleeyes.log 2>/dev/null
rm tendermint.log 2>/dev/null
rm query.txt 2>/dev/null
