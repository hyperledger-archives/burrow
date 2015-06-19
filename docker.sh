#! /bin/bash

docker build -t eris-db .

# run eris-db 
docker run --name eris-db -p 46656:46656 -p 46657:46657 -p 1337:1337 eris-db
