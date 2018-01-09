#!/bin/bash

./start-tendermerk.sh
./transaction.sh
./height.sh
./transaction.sh
./height.sh
./transaction.sh
./height.sh
./kill-tendermerk.sh

./restart-tendermerk.sh
./transaction.sh
./height.sh
./transaction.sh
./height.sh
./kill-tendermerk.sh
