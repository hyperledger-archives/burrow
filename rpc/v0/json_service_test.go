// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v0

import (
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-wire"
)

func TestBroadcastTx(t *testing.T) {
	testData := LoadTestData()
	pipe := NewMockPipe(testData)
	methods := NewBurrowMethods(NewTCodec(), pipe)
	pubKey := acm.GeneratePrivateAccount().PubKey()
	address := acm.Address{1}
	code := bc.Splice(asm.PUSH1, 1, asm.PUSH1, 1, asm.ADD)
	var tx txs.Tx = txs.NewCallTxWithNonce(pubKey, &address, code, 10, 2,
		1, 0)
	jsonBytes := wire.JSONBytesPretty(wrappedTx{tx})
	request := rpc.NewRPCRequest("TestBroadcastTx", "BroadcastTx", jsonBytes)
	result, _, err := methods.BroadcastTx(request, "TestBroadcastTx")
	assert.NoError(t, err)
	receipt, ok := result.(*txs.Receipt)
	assert.True(t, ok, "Should get Receipt pointer")
	assert.Equal(t, txs.TxHash(testData.GetChainId.Output.ChainId, tx), receipt.TxHash)
}

// Allows us to get the type byte included but then omit the outer struct and
// embedded field
type wrappedTx struct {
	txs.Tx `json:"unwrap"`
}
