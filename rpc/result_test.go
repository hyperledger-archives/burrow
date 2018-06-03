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

package rpc

import (
	"encoding/json"
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-wire"
	tm_types "github.com/tendermint/tendermint/types"
)

func TestResultBroadcastTx(t *testing.T) {
	// Make sure these are unpacked as expected
	res := ResultBroadcastTx{
		Receipt: txs.Receipt{
			ContractAddress: crypto.Address{0, 2, 3},
			CreatesContract: true,
			TxHash:          []byte("foo"),
		},
	}

	js := string(wire.JSONBytes(res))
	assert.Equal(t, `{"Receipt":{"TxHash":"666F6F","CreatesContract":true,"ContractAddress":"0002030000000000000000000000000000000000"}}`, js)

	res2 := new(ResultBroadcastTx)
	wire.ReadBinaryBytes(wire.BinaryBytes(res), res2)
	assert.Equal(t, res, *res2)
}

func TestListUnconfirmedTxs(t *testing.T) {
	res := &ResultListUnconfirmedTxs{
		NumTxs: 3,
		Txs: []txs.Wrapper{
			txs.Wrap(&txs.CallTx{
				Address: &crypto.Address{1},
			}),
		},
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	assert.Equal(t, `{"NumTxs":3,"Txs":[{"type":"call_tx","data":{"Input":null,"Address":"0100000000000000000000000000000000000000","GasLimit":0,"Fee":0,"Data":null}}]}`,
		string(bs))
}

func TestResultListAccounts(t *testing.T) {
	concreteAcc := acm.AsConcreteAccount(acm.FromAddressable(
		acm.GeneratePrivateAccountFromSecret("Super Semi Secret")))
	acc := concreteAcc
	res := ResultListAccounts{
		Accounts:    []*acm.ConcreteAccount{acc},
		BlockHeight: 2,
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)
	resOut := new(ResultListAccounts)
	json.Unmarshal(bs, resOut)
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

func TestResultCall_MarshalJSON(t *testing.T) {
	res := ResultCall{
		Call: execution.Call{
			Return:  []byte("hi"),
			GasUsed: 1,
		},
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)

	resOut := new(ResultCall)
	json.Unmarshal(bs, resOut)
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

func TestResultEvent(t *testing.T) {
	eventDataNewBlock := tm_types.EventDataNewBlock{
		Block: &tm_types.Block{
			Header: &tm_types.Header{
				ChainID: "chainy",
				NumTxs:  30,
			},
		},
	}
	res := ResultEvent{
		Tendermint: &ResultTendermintEvent{
			EventDataNewBlock: &eventDataNewBlock,
		},
	}
	bs, err := json.Marshal(res)
	require.NoError(t, err)

	resOut := new(ResultEvent)
	json.Unmarshal(bs, resOut)
	bsOut, err := json.Marshal(resOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
	//fmt.Println(string(bs))
	//fmt.Println(string(bsOut))
}
