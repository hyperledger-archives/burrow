#!/bin/bash

echo "Sending a transaction"

echo -n "Hash="
tx=`./gen_merk_trans.sh`

curl http://localhost:46657/broadcast_tx_commit\?tx\=${tx} 2>/dev/null |\
jq 'if .result == 0 then
  "No Result"
elif .check_tx.log > 2 then
  .result | .check_tx.log
else
  .result | .hash
end'

echo ""
