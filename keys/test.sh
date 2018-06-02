#! /bin/bash

# tests
# run the suite with and without the daemon

# TODO: run the suite with/without encryption!


burrow_bin=${burrow_bin:-burrow}

echo "-----------------------------"
echo "starting the server"
$burrow_bin keys server &
keys_pid=$!
sleep 1
echo "-----------------------------"
echo "testing the cli"

# we test keys, hashes, names, and import
# this file should be run with and without the daemon running

echo "testing keys"

CURVETYPES=("ed25519" "secp256k1")
for CURVETYPE in ${CURVETYPES[*]}
do
	# test key gen, sign verify:
	# for each step, ensure it works using --addr or --name
	echo "... $CURVETYPE"

	HASH=`$burrow_bin keys hash --type sha256 ok`
	#echo "HASH: $HASH"
	NAME=testkey1
	ADDR=`$burrow_bin keys gen --curvetype $CURVETYPE --name $NAME --no-password`
	#echo "my addr: $ADDR"
	PUB1=`$burrow_bin keys pub --name $NAME`
	PUB2=`$burrow_bin keys pub --addr $ADDR`
	if [ "$PUB1" != "$PUB2" ]; then
		echo "FAILED pub: got $PUB2, expected $PUB1"
		kill $keys_pid
		exit 1
	fi
	echo "...... passed pub"

	SIG1=`$burrow_bin keys sign --name $NAME $HASH`
	VERIFY1=`$burrow_bin keys verify --curvetype $CURVETYPE $HASH $SIG1 $PUB1`
	if [ $VERIFY1 != "true" ]; then
		echo "FAILED verify: got $VERIFY1 expected true"
		kill $keys_pid
		exit 1
	fi

	SIG2=`$burrow_bin keys sign --addr $ADDR $HASH`
	VERIFY1=`$burrow_bin keys verify --curvetype $CURVETYPE $HASH $SIG2 $PUB1`
	if [ $VERIFY1 != "true" ]; then
		echo "FAILED verify: got $VERIFY1 expected true"
		kill $keys_pid
		exit 1
	fi

	echo "...... passed sig/verify"

done

echo "testing hashes"
# test hashing (we need openssl)
TOHASH=okeydokey
HASHTYPES=(sha256 ripemd160)
for HASHTYPE in ${HASHTYPES[*]}
do
	echo  "... $HASHTYPE"
	HASH0=`echo -n $TOHASH | openssl dgst -$HASHTYPE | awk '{print toupper($2)}'`
	HASH1=`$burrow_bin keys hash --type $HASHTYPE $TOHASH`
	if [ "$HASH0" != "$HASH1" ]; then
		echo "FAILED hash $HASHTYPE: got $HASH1 expected $HASH0"
	fi
	echo "...... passed"
done

echo "testing imports"

# TODO: IMPORTS
# for each key type, import a priv key, ensure it returns
# the right address. do again with both plain and encrypted jsons

for CURVETYPE in ${CURVETYPES[*]}
do
	echo "... $CURVETYPE"
	# create a key, get its address and priv, backup the json, delete the key
	ADDR=`$burrow_bin keys gen --curvetype $CURVETYPE --no-password`
	DIR=.keys/data
	FILE=$DIR/$ADDR.json
	PRIV=`cat $FILE |  jq -r .PrivateKey.Plain`
	HEXPRIV=`echo -n "$PRIV" | base64 -d | xxd -p -u -c 256`
	cp $FILE ~/$ADDR
	rm -rf $DIR

	# import the key via priv
	ADDR2=`$burrow_bin keys import --no-password --curvetype $CURVETYPE $HEXPRIV`
	if [ "$ADDR" != "$ADDR2" ]; then
		echo "FAILED import $CURVETYPE: got $ADDR2 expected $ADDR"
		kill $keys_pid
		exit
	fi
	rm -rf $DIR

	# import the key via json
	JSON=`cat ~/$ADDR`
	ADDR2=`$burrow_bin keys import --no-password --curvetype $CURVETYPE $JSON`
	if [ "$ADDR" != "$ADDR2" ]; then
		echo "FAILED import (json) $CURVETYPE: got $ADDR2 expected $ADDR"
		kill $keys_pid
		exit
	fi
	rm -rf $DIR

	# import the key via path
	ADDR2=`$burrow_bin keys import --no-password --curvetype $CURVETYPE ~/$ADDR`
	if [ "$ADDR" != "$ADDR2" ]; then
		echo "FAILED import $CURVETYPE: got $ADDR2 expected $ADDR"
		kill $keys_pid
		exit
	fi
	rm -rf $DIR

	echo "...... passed raw hex and json"
done


echo "testing names"

NAME=mykey
ADDR=`$burrow_bin keys gen --name $NAME --no-password`
ADDR2=`$burrow_bin keys list --name $NAME`
if [ "$ADDR" != "$ADDR2" ]; then
	echo "FAILED name: got $ADDR2 expected $ADDR"
	kill $keys_pid
	exit
fi

NAME2=mykey2
$burrow_bin keys name $NAME2 $ADDR
ADDR2=`$burrow_bin keys list --name $NAME2`
if [ "$ADDR" != "$ADDR2" ]; then
	echo "FAILED rename: got $ADDR2 expected $ADDR"
	kill $keys_pid
	exit
fi

echo "... passed"

kill $keys_pid

# TODO a little more on names...

