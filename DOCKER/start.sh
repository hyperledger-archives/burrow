#!/bin/bash
if [[ $FAST_SYNC ]]; then
  tendermint node --fast_sync
else
  tendermint node
fi

