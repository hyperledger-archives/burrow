#! /bin/bash
set -e

NAME="net_test"

ADDR=`mintkey eris data/${NAME}_0/priv_validator.json`
MINTX_PUBKEY=""
echo $ADDR

mintx send --to=1A73F65F86271552C4858091D3DADFE3335B9273 --addr=$ADDR --amt=1 --sign --broadcast --log=6 --chainID=$NAME
