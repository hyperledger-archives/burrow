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
	"fmt"
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var txEnvelopeString = `{"Signatories":[{"Address":"83207817DC3814B96F57EFF925F467E07CAA9138","PublicKey":{"CurveType":"ed25519",` +
	`"PublicKey":"34D26579DBB456693E540672CF922F52DDE0D6532E35BF06BE013A7C532F20E0"},` +
	`"Signature":"5042F208824AA5AF8E03B2F11FB8CFCDDAE4F889B2F720714627395406E00D7740B2DB5B5F93BD6C13DED9B7C1FD5FB0DB4ECA31E6DA0B81033A72922076E90C"}],` +
	`"Tx":{"ChainID":"testChain","Type":"CallTx","Payload":{"Input":{"Address":"83207817DC3814B96F57EFF925F467E07CAA9138","Amount":343,"Sequence":3},` +
	`"Address":"AC280D53FD359D9FF11F19D0796D9B89907F3B53","GasLimit":2323,"Fee":12,"Data":"03040505"}}}`

var testBroadcastCallTxJsonRequest = []byte(`
{
  "id": "57EC1D39-7B3D-4F96-B286-8FC128177AFC4",
  "jsonrpc": "2.0",
  "method": "burrow.broadcastTx",
  "params": ` + txEnvelopeString + `
}`)

// strictly test the codec for go-wire encoding of the Json format,
// This should restore compatibility with the format on v0.11.4
// (which was broken on v0.12)
func fixTestCallTxJsonFormatCodec(t *testing.T) {
	codec := NewTCodec()
	txEnv := new(txs.Envelope)
	// Create new request object and unmarshal.
	request := &rpc.RPCRequest{}
	assert.NoError(t, json.Unmarshal(testBroadcastCallTxJsonRequest, request),
		"Provided JSON test data does not unmarshal to rpc.RPCRequest object.")
	assert.NoError(t, codec.DecodeBytes(txEnv, request.Params),
		"RPC codec failed to decode params as transaction type.")
	_, ok := txEnv.Tx.Payload.(*payload.CallTx)
	assert.True(t, ok, "Type byte 0x02 should unmarshal into CallTx.")
}

func TestGenTxEnvelope(t *testing.T) {
	codec := NewTCodec()
	privAccFrom := acm.GeneratePrivateAccountFromSecret("foo")
	privAccTo := acm.GeneratePrivateAccountFromSecret("bar")
	toAddress := privAccTo.Address()
	txEnv := txs.Enclose("testChain", payload.NewCallTxWithSequence(privAccFrom.PublicKey(), &toAddress,
		[]byte{3, 4, 5, 5}, 343, 2323, 12, 3))
	err := txEnv.Sign(privAccFrom)
	require.NoError(t, err)
	bs, err := codec.EncodeBytes(txEnv)
	require.NoError(t, err)
	if !assert.Equal(t, txEnvelopeString, string(bs)) {
		fmt.Println(string(bs))
	}
}
