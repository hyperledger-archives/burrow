#!/bin/bash

if [ "$1" = "" ]; then
	key=$(go run eightbytekey.go)
else 
	key=$1
fi
key_size=0108
value=33
value_size=0101

echo -n 0x01${key_size}${key}${value_size}${value}

