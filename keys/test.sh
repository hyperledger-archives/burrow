#! /bin/bash

# tests
# run the suite with and without the daemon

# TODO: run the suite with/without encryption!

set -e

burrow_bin=${burrow_bin:-burrow}

test_dir="./keys/test_scratch"
keys_dir="$test_dir/.keys"
tmp_dir="$test_dir/tmp"
mkdir -p "$tmp_dir"

echo "-----------------------------"
echo "checking for dependent utilities"
for UTILITY in jq xxd openssl; do
    echo -n "... "
    if ! command -v $UTILITY; then
        echo "$UTILITY (missing)"
        missing_utility=$UTILITY
    fi
done
if [ ! -z $missing_utility ]; then
    echo "FAILED dependency check: the '$missing_utility' utility is missing"
    exit 1
fi

echo "starting the server"
$burrow_bin keys server --dir $keys_dir &
keys_pid=$!
function kill_burrow_keys {
    kill -TERM $keys_pid
}

trap kill_burrow_keys EXIT
sleep 1
echo "-----------------------------"
echo "testing the cli"

# we test keys, hashes, names, and import
# this file should be run with and without the daemon running

echo "testing keys"

CURVETYPES=( "ed25519" "secp256k1" )
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
        exit 1
    fi
    echo "...... passed pub"

    SIG1=`$burrow_bin keys sign --name $NAME $HASH`
    VERIFY1=`$burrow_bin keys verify --curvetype $CURVETYPE $HASH $SIG1 $PUB1`
    if [ $VERIFY1 != "true" ]; then
        echo "FAILED verify: got $VERIFY1 expected true"
        exit 1
    fi

    SIG2=`$burrow_bin keys sign --addr $ADDR $HASH`
    VERIFY1=`$burrow_bin keys verify --curvetype $CURVETYPE $HASH $SIG2 $PUB1`
    if [ $VERIFY1 != "true" ]; then
        echo "FAILED verify: got $VERIFY1 expected true"
        exit 1
    fi

    echo "...... passed sig/verify"

done

echo "testing hashes"
# test hashing (we need openssl)
TOHASH=okeydokey
HASHTYPES=( sha256 ripemd160 )
for HASHTYPE in ${HASHTYPES[*]}
do
    echo "... $HASHTYPE"
    # XXX: OpenSSL's `openssl dgst -<hash>` command might produce both
    # a one-field (LibreSSL 2.2.7)
    #
    #   $ echo -n okeydokey |openssl dgst -sha256
    #   0fd2479fa22057f562698c4e6bb5b6c7430a10ba0fe6cd41fa9908e2c0a684a4
    #
    # and a two-field result (OpenSSL 1.1.0f):
    #
    #   $ echo -n okeydokey |openssl dgst -sha256
    #   (stdin)= 0fd2479fa22057f562698c4e6bb5b6c7430a10ba0fe6cd41fa9908e2c0a684a4
    #
    # Generalize to adjust for the inconsistency:
    HASH0=`echo -n $TOHASH | openssl dgst -$HASHTYPE | sed 's/^.* //' | tr '[:lower:]' '[:upper:]'`
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
    DIR=$keys_dir/data
    FILE=$DIR/$ADDR.json
    HEXPRIV=`cat $FILE | jq -r .PrivateKey.Plain`
    EXPORTJSON=`$burrow_bin keys export --addr $ADDR`


    cp $FILE "$tmp_dir/$ADDR"
    rm -rf $DIR

    # import the key via priv
    ADDR2=`$burrow_bin keys import --no-password --curvetype $CURVETYPE $HEXPRIV`
    if [ "$ADDR" != "$ADDR2" ]; then
        echo "FAILED import $CURVETYPE: got $ADDR2 expected $ADDR"
        exit
    fi
    rm -rf $DIR

    # import the key via json
    JSON=`cat "$tmp_dir/$ADDR"`
    ADDR2=`$burrow_bin keys import --no-password --curvetype $CURVETYPE $JSON`
    if [ "$ADDR" != "$ADDR2" ]; then
        echo "FAILED import (json) $CURVETYPE: got $ADDR2 expected $ADDR"
        exit
    fi
    rm -rf $DIR

    # import the key via path
    ADDR2=`$burrow_bin keys import --no-password --curvetype $CURVETYPE "$tmp_dir/$ADDR"`
    if [ "$ADDR" != "$ADDR2" ]; then
        echo "FAILED import $CURVETYPE: got $ADDR2 expected $ADDR"
        exit
    fi
    rm -rf $DIR

    # import the key via export json
    ADDR2=`$burrow_bin keys import --no-password "$EXPORTJSON"`
    if [ "$ADDR" != "$ADDR2" ]; then
        echo "FAILED import from export $CURVETYPE: got $ADDR2 expected $ADDR"
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
    exit
fi

NAME2=mykey2
$burrow_bin keys name $NAME2 $ADDR
ADDR2=`$burrow_bin keys list --name $NAME2`
if [ "$ADDR" != "$ADDR2" ]; then
    echo "FAILED rename: got $ADDR2 expected $ADDR"
    exit
fi

echo "... passed"

rm -rf "$test_dir"

# TODO a little more on names...

