#! /bin/bash
set -e

# tests
# run the suite with and without the daemon

# TODO: run the suite with/without encryption!

echo "-----------------------------"
echo "starting the server"
monax-keys server &
sleep 1
echo "-----------------------------"
echo "testing the cli"

# we test keys, hashes, names, and import
# this file should be run with and without the daemon running

echo "testing keys"

KEYTYPES=("ed25519,ripemd160" "secp256k1,sha3" "secp256k1,ripemd160sha256")
for KEYTYPE in ${KEYTYPES[*]}
do
	# test key gen, sign verify:
	# for each step, ensure it works using --addr or --name
	echo "... $KEYTYPE"

	HASH=`monax-keys hash --type sha256 ok`
	#echo "HASH: $HASH"
	NAME=testkey1
	ADDR=`monax-keys gen --type $KEYTYPE --name $NAME --no-pass`
	#echo "my addr: $ADDR"
	PUB1=`monax-keys pub --name $NAME`
	PUB2=`monax-keys pub --addr $ADDR`
	if [ "$PUB1" != "$PUB2" ]; then
		echo "FAILED pub: got $PUB2, expected $PUB1"
		exit 1
	fi
	echo "...... passed pub"

	SIG1=`monax-keys sign --name $NAME $HASH`
	VERIFY1=`monax-keys verify --type $KEYTYPE $HASH $SIG1 $PUB1`
	if [ $VERIFY1 != "true" ]; then
		echo "FAILED verify: got $VERIFY1 expected true"
		exit 1
	fi

	SIG2=`monax-keys sign --addr $ADDR $HASH`
	VERIFY1=`monax-keys verify --type $KEYTYPE $HASH $SIG2 $PUB1`
	if [ $VERIFY1 != "true" ]; then
		echo "FAILED verify: got $VERIFY1 expected true"
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
	HASH1=`monax-keys hash --type $HASHTYPE $TOHASH`
	if [ "$HASH0" != "$HASH1" ]; then
		echo "FAILED hash $HASHTYPE: got $HASH1 expected $HASH0"
	fi
	echo "...... passed"
done

echo "testing imports"

# TODO: IMPORTS
# for each key type, import a priv key, ensure it returns
# the right address. do again with both plain and encrypted jsons

for KEYTYPE in ${KEYTYPES[*]}
do
	echo "... $KEYTYPE"
	# create a key, get its address and priv, backup the json, delete the key
	ADDR=`monax-keys gen --type $KEYTYPE --no-pass`
	DIR=$HOME/.monax/keys/data/$ADDR
	FILE=$DIR/$ADDR
	PRIV=`cat $FILE |  jq -r .PrivateKey`
	HEXPRIV=`echo -n "$PRIV" | base64 -d | hexdump -ve '1/1 "%.2X"'`
	cp $FILE ~/$ADDR
	rm -rf $DIR

	# import the key via priv
	ADDR2=`monax-keys import --no-pass --type $KEYTYPE $HEXPRIV`
	if [ "$ADDR" != "$ADDR2" ]; then
		echo "FAILED import $KEYTYPE: got $ADDR2 expected $ADDR"	
		exit
	fi
	rm -rf $DIR

	# import the key via json
	JSON=`cat ~/$ADDR`
	ADDR2=`monax-keys import --no-pass --type $KEYTYPE $JSON`
	if [ "$ADDR" != "$ADDR2" ]; then
		echo "FAILED import (json) $KEYTYPE: got $ADDR2 expected $ADDR"	
		exit
	fi
	rm -rf $DIR

	# import the key via path
	ADDR2=`monax-keys import --no-pass --type $KEYTYPE ~/$ADDR`
	if [ "$ADDR" != "$ADDR2" ]; then
		echo "FAILED import $KEYTYPE: got $ADDR2 expected $ADDR"	
		exit
	fi
	rm -rf $DIR

	echo "...... passed raw hex and json"
done


echo "testing names"

NAME=mykey
ADDR=`monax-keys gen --name $NAME --no-pass`
ADDR2=`monax-keys name $NAME`
if [ "$ADDR" != "$ADDR2" ]; then
	echo "FAILED name: got $ADDR2 expected $ADDR"	
	exit
fi

NAME2=mykey2
monax-keys name $NAME2 $ADDR
ADDR2=`monax-keys name $NAME2`
if [ "$ADDR" != "$ADDR2" ]; then
	echo "FAILED rename: got $ADDR2 expected $ADDR"	
	exit
fi

echo "... passed"


# TODO a little more on names...

