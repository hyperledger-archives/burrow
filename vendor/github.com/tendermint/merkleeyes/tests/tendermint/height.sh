#!/bin/bash

echo "Sending a transaction"

key=$(go run eightbytekey.go)
tx=`./gen_merk_trans.sh ${key}`

echo -n "Hash="
curl http://localhost:46657/broadcast_tx_commit\?tx\=${tx} 2>/dev/null |\
jq 'if .result == 0 then
  "No Result"
elif .check_tx.log > 2 then
  .result | .check_tx.log
else
  .result | .hash
end'
echo ""

echo "Checking its height"

curl http://localhost:46657/status 2>/dev/null > status.txt
curl http://localhost:46657/abci_query\?data\=0x${key}\&path\=\"/key\"\&prove=true 2>/dev/null > query.txt

echo -n "status height="
cat status.txt | jq '.result.latest_block_height'

echo -n "query height="
cat query.txt | jq '.result.response.height'

echo -n "status apphash="
cat status.txt | jq '.result.latest_app_hash'

# get the apphash from the proof
echo -n "query proof="
cat query.txt | jq '.result.response.proof'

# A second time
sleep 1
curl http://localhost:46657/status 2>/dev/null > status.txt
curl http://localhost:46657/abci_query\?data\=0x${key}\&path\=\"/key\"\&prove=true 2>/dev/null > query.txt

echo -n "status height="
cat status.txt | jq '.result.latest_block_height'

echo -n "query height="
cat query.txt | jq '.result.response.height'

echo -n "status apphash="
cat status.txt | jq '.result.latest_app_hash'

# get the apphash from the proof
echo -n "query proof="
cat query.txt | jq '.result.response.proof' 

echo ""
