#! /bin/bash
set -e

NAME="net_test"

ADDR=`mintkey eris data/${NAME}_0/priv_validator.json`
MINTX_PUBKEY=""
echo $ADDR

TO=1A73F65F86271552C4858091D3DADFE3335B9273 

set +e
START_BALANCE=`mintinfo accounts $TO`
if [[ "$?" == 1 ]]; then
	START_BALANCE=0

else
	START_BALANCE=`echo $START_BALANCE | jq .balance`
fi
set -e

N=50
for i in `seq 1 $N`; do
	mintx send --to=$TO --addr=$ADDR --amt=1 --sign --broadcast --chainID=$NAME
done
mintx send --to=$TO --addr=$ADDR --amt=1 --sign --broadcast --chainID=$NAME --log=6 --wait

END_BALANCE=`mintinfo accounts $TO | jq .balance`

DIFF_BALANCE=$((${END_BALANCE}-${START_BALANCE}))

if [[ "$DIFF_BALANCE" != $(($N+1)) ]]; then
	echo "Failed! Expected $((N+1)), got $DIFF_BALANCE"
	exit 1
fi

echo "PASS"


