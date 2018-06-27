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
	"sync"
	"testing"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/rpc/test"
	"github.com/hyperledger/burrow/rpc/v0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactCallNoCode(t *testing.T) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")

	// Flip flops between sending private key and input address to test private key and address based signing
	toAddress := privateAccounts[2].Address()

	numCreates := 1000
	countCh := test.CommittedTxCount(t, kern.Emitter)
	for i := 0; i < numCreates; i++ {
		receipt, err := cli.Transact(v0.TransactParam{
			InputAccount: inputAccount(i),
			Address:      toAddress.Bytes(),
			Data:         []byte{},
			Fee:          2,
			GasLimit:     10000 + uint64(i),
		})
		require.NoError(t, err)
		assert.False(t, receipt.CreatesContract)
		assert.Equal(t, toAddress, receipt.ContractAddress)
	}
	require.Equal(t, numCreates, <-countCh)
}

func TestTransactCreate(t *testing.T) {
	numGoroutines := 100
	numCreates := 50
	wg := new(sync.WaitGroup)
	wg.Add(numGoroutines)
	cli := v0.NewV0Client("http://localhost:1337/rpc")
	// Flip flops between sending private key and input address to test private key and address based signing
	countCh := test.CommittedTxCount(t, kern.Emitter)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numCreates; j++ {
				create, err := cli.Transact(v0.TransactParam{
					InputAccount: inputAccount(i),
					Address:      nil,
					Data:         test.Bytecode_strange_loop,
					Fee:          2,
					GasLimit:     10000,
				})
				if assert.NoError(t, err) {
					assert.True(t, create.CreatesContract)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	require.Equal(t, numGoroutines*numCreates, <-countCh)
}

func BenchmarkTransactCreateContract(b *testing.B) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")

	for i := 0; i < b.N; i++ {
		create, err := cli.Transact(v0.TransactParam{
			InputAccount: inputAccount(i),
			Address:      nil,
			Data:         test.Bytecode_strange_loop,
			Fee:          2,
			GasLimit:     10000,
		})
		require.NoError(b, err)
		assert.True(b, create.CreatesContract)
	}
}

func TestTransactAndHold(t *testing.T) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")

	numGoroutines := 5
	numRuns := 2
	countCh := test.CommittedTxCount(t, kern.Emitter)
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numRuns; j++ {
			create, err := cli.TransactAndHold(v0.TransactParam{
				InputAccount: inputAccount(i),
				Address:      nil,
				Data:         test.Bytecode_strange_loop,
				Fee:          2,
				GasLimit:     10000,
			})
			require.NoError(t, err)
			assert.Equal(t, uint64(0), create.StackDepth)
			functionID := abi.FunctionID("UpsieDownsie()")
			call, err := cli.TransactAndHold(v0.TransactParam{
				InputAccount: inputAccount(i),
				Address:      create.CallData.Callee.Bytes(),
				Data:         functionID[:],
				Fee:          2,
				GasLimit:     10000,
			})
			require.NoError(t, err)
			depth := binary.Uint64FromWord256(binary.LeftPadWord256(call.Return))
			// Would give 23 if taken from wrong frame
			assert.Equal(t, 18, int(depth))
		}
	}
	require.Equal(t, numGoroutines*numRuns*2, <-countCh)
}

func TestSend(t *testing.T) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")
	numSends := 1000
	countCh := test.CommittedTxCount(t, kern.Emitter)
	for i := 0; i < numSends; i++ {
		send, err := cli.Send(v0.SendParam{
			InputAccount: inputAccount(i),
			Amount:       2003,
			ToAddress:    privateAccounts[3].Address().Bytes(),
		})
		require.NoError(t, err)
		assert.Equal(t, false, send.CreatesContract)
	}
	require.Equal(t, numSends, <-countCh)
}

func TestSendAndHold(t *testing.T) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")

	for i := 0; i < 2; i++ {
		send, err := cli.SendAndHold(v0.SendParam{
			InputAccount: inputAccount(i),
			Amount:       2003,
			ToAddress:    privateAccounts[3].Address().Bytes(),
		})
		require.NoError(t, err)
		assert.Equal(t, false, send.CreatesContract)
	}
}

// Helpers

var inputPrivateKey = privateAccounts[0].PrivateKey().RawBytes()
var inputAddress = privateAccounts[0].Address().Bytes()

func inputAccount(i int) v0.InputAccount {
	ia := v0.InputAccount{}
	if i%2 == 0 {
		ia.PrivateKey = inputPrivateKey
	} else {
		ia.Address = inputAddress
	}
	return ia
}
