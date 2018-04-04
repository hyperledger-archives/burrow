// +build integration

// Space above here matters
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

package integration

import (
	"testing"

	"github.com/hyperledger/burrow/rpc/v0"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransact(t *testing.T) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")
	receipt := new(txs.Receipt)

	address := privateAccounts[1].Address()
	param := v0.TransactParam{
		PrivKey:  privateAccounts[0].PrivateKey().RawBytes(),
		Address:  address.Bytes(),
		Data:     []byte{},
		Fee:      2,
		GasLimit: 10000,
	}
	err := cli.Call(v0.TRANSACT, param, receipt)
	require.NoError(t, err)
	assert.False(t, receipt.CreatesContract)
	assert.Equal(t, address, receipt.ContractAddress)
}
