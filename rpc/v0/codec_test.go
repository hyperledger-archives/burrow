// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v0

import (
	"encoding/json"
	"testing"

	"github.com/eris-ltd/eris-db/txs"

	"github.com/eris-ltd/eris-db/rpc"
	"github.com/stretchr/testify/assert"
)

var testBroadcastCallTxJsonRequest = []byte(`
{
  "id": "57EC1D39-7B3D-4F96-B286-8FC128177AFC4",
  "jsonrpc": "2.0",
  "method": "erisdb.broadcastTx",
  "params": [
    2,
    {
      "address": "5A9083BB0EFFE4C8EB2ADD29174994F73E77D418",
      "data": "2F2397A00000000000000000000000000000000000000000000000000000000000003132",
      "fee": 1,
      "gas_limit": 1000000,
      "input": {
        "address": "BE18FDCBF12BF99F4D75325E17FF2E78F1A35FE8",
        "amount": 1,
        "pub_key": [
          1,
          "8D1611925948DC2EDDF739FB65CE517757D286155A039B28441C3349BE9A8C38"
        ],
        "sequence": 2,
        "signature": [
          1,
          "B090D622F143ECEDA9B9E7B15485CE7504453C05434951CF867B013D80ED1BD2A0CA32846FC175D234CDFB9D5C3D792759E8FE79FD4DB3006B24950EE3C37D00"
        ]
      }
    }
  ]
}`)

// strictly test the codec for go-wire encoding of the Json format,
// This should restore compatibility with the format on v0.11.4
// (which was broken on v0.12)
func TestCallTxJsonFormatCodec(t *testing.T) {
	codec := NewTCodec()
	param := new(txs.Tx)

	// Create new request object and unmarshal.
	request := &rpc.RPCRequest{}
	assert.NoError(t, json.Unmarshal(testBroadcastCallTxJsonRequest, request),
		"Provided JSON test data does not unmarshal to rpc.RPCRequest object.")
	assert.NoError(t, codec.DecodeBytes(param, request.Params),
		"RPC codec failed to decode params as transaction type.")
	_, ok := (*param).(*txs.CallTx)
	assert.True(t, ok, "Type byte 0x02 should unmarshal into CallTx.")
}
