#!/bin/bash

# Using ~/.eris on drive.
ERIS_PATH=$HOME/.eris
CONTAINER="eris-db"
RUNNING=$(docker inspect --format="{{ .State.Running }}" eris-db)
mkdir -v -p $ERIS_PATH

# Run in the terminal and attach on start.
if [ "$RUNNING" == "true" ]; then
  echo "Container 'eris-db' already running. Exiting."
  exit 1
elif [ "$RUNNING" == "false" ]; then
  echo "Container 'eris-db' found. Starting."
  docker start --attach=true eris-db
else
  echo "Container 'eris-db' not found. Creating."
  docker run --name eris-db -v $ERIS_PATH:/home/eris/.eris -p 46656:46656 -p 46657:46657 -p 1337:1337 eris-db
fi
