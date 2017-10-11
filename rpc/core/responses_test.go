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

package core

import (
	"testing"

	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-wire"
	"encoding/json"
)

func TestResultBroadcastTx(t *testing.T) {
	// Make sure these are unpacked as expected
	res := ResultBroadcastTx{
		Receipt: txs.Receipt{
			ContractAddr:    acm.Address{0, 2, 3},
			CreatesContract: true,
			TxHash:          []byte("foo"),
		},
	}

	assert.Equal(t, `{"tx_hash":"666F6F","creates_contract":true,"contract_addr":"0002030000000000000000000000000000000000"}`,
		string(wire.JSONBytes(res)))

	res2 := new(ResultBroadcastTx)
	wire.ReadBinaryBytes(wire.BinaryBytes(res), res2)
	assert.Equal(t, res, *res2)
}

func TestListUnconfirmedTxs(t *testing.T) {
	res := &ResultListUnconfirmedTxs{
		N: 3,
		Txs: []txs.Tx{
			&txs.CallTx{
				Address: &acm.Address{1},
			},
		},
	}
	fmt.Println(string(wire.JSONBytes(res)))

}

func TestJSONEncode(t *testing.T) {
	concreteAcc := acm.AsConcreteAccount(acm.FromAddressable(
		acm.GeneratePrivateAccountFromSecret("Super Semi Secret")))
	acc := concreteAcc.Account()
	res := &ResultListAccounts{
		Accounts:    []acm.Account{acc},
		BlockHeight: 2,
	}
	fmt.Println(string(wire.JSONBytes(res)))
	fmt.Println("")
	bs, _ := json.Marshal(res)
	fmt.Println(string(bs))
}
