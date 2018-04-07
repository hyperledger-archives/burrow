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
	"encoding/hex"
	"testing"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/rpc/v0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactCallNoCode(t *testing.T) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")

	// Flip flops between sending private key and input address to test private key and address based signing
	privKey, inputAddress := privKeyInputAddressAlternator(privateAccounts[0])
	toAddress := privateAccounts[2].Address()

	for i := 0; i < 1000; i++ {
		receipt, err := cli.Transact(v0.TransactParam{
			PrivKey:      privKey(i),
			InputAddress: inputAddress(i),
			Address:      toAddress.Bytes(),
			Data:         []byte{},
			Fee:          2,
			GasLimit:     10000 + uint64(i),
		})
		require.NoError(t, err)
		assert.False(t, receipt.CreatesContract)
		assert.Equal(t, toAddress, receipt.ContractAddress)
	}
}

func TestTransactCreate(t *testing.T) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")

	// Flip flops between sending private key and input address to test private key and address based signing
	privKey, inputAddress := privKeyInputAddressAlternator(privateAccounts[0])
	for i := 0; i < 1000; i++ {
		bc, err := hex.DecodeString(strangeLoopBytecode)
		require.NoError(t, err)
		create, err := cli.Transact(v0.TransactParam{
			PrivKey:      privKey(i),
			InputAddress: inputAddress(i),
			Address:      nil,
			Data:         bc,
			Fee:          2,
			GasLimit:     10000,
		})
		require.NoError(t, err)
		assert.True(t, create.CreatesContract)
	}
}

func BenchmarkTransactCreateContract(b *testing.B) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")

	privKey, inputAddress := privKeyInputAddressAlternator(privateAccounts[0])
	for i := 0; i < b.N; i++ {
		bc, err := hex.DecodeString(strangeLoopBytecode)
		require.NoError(b, err)
		create, err := cli.Transact(v0.TransactParam{
			PrivKey:      privKey(i),
			InputAddress: inputAddress(i),
			Address:      nil,
			Data:         bc,
			Fee:          2,
			GasLimit:     10000,
		})
		require.NoError(b, err)
		assert.True(b, create.CreatesContract)
	}
}

func TestTransactAndHold(t *testing.T) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")

	bc, err := hex.DecodeString(strangeLoopBytecode)
	require.NoError(t, err)

	privKey, inputAddress := privKeyInputAddressAlternator(privateAccounts[0])

	for i := 0; i < 2; i++ {
		create, err := cli.TransactAndHold(v0.TransactParam{
			PrivKey:      privKey(i),
			InputAddress: inputAddress(i),
			Address:      nil,
			Data:         bc,
			Fee:          2,
			GasLimit:     10000,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, create.StackDepth)
		functionID := abi.FunctionID("UpsieDownsie()")
		call, err := cli.TransactAndHold(v0.TransactParam{
			PrivKey:      privKey(i),
			InputAddress: inputAddress(i),
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

func TestSend(t *testing.T) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")

	privKey, inputAddress := privKeyInputAddressAlternator(privateAccounts[0])
	for i := 0; i < 1000; i++ {
		send, err := cli.Send(v0.SendParam{
			PrivKey:      privKey(i),
			InputAddress: inputAddress(i),
			Amount:       2003,
			ToAddress:    privateAccounts[3].Address().Bytes(),
		})
		require.NoError(t, err)
		assert.Equal(t, false, send.CreatesContract)
	}
}

func TestSendAndHold(t *testing.T) {
	cli := v0.NewV0Client("http://localhost:1337/rpc")

	privKey, inputAddress := privKeyInputAddressAlternator(privateAccounts[0])

	for i := 0; i < 2; i++ {
		send, err := cli.SendAndHold(v0.SendParam{
			PrivKey:      privKey(i),
			InputAddress: inputAddress(i),
			Amount:       2003,
			ToAddress:    privateAccounts[3].Address().Bytes(),
		})
		require.NoError(t, err)
		assert.Equal(t, false, send.CreatesContract)
	}
}

// Returns a pair of functions that mutually exclusively return the private key bytes or input address bytes of a
// private account in the same iteration of a loop indexed by an int
func privKeyInputAddressAlternator(privateAccount account.PrivateAccount) (func(int) []byte, func(int) []byte) {
	privKey := privateAccount.PrivateKey().RawBytes()
	inputAddress := privateAccount.Address().Bytes()
	return alternator(privKey, 0), alternator(inputAddress, 1)
}

func alternator(ret []byte, res int) func(int) []byte {
	return func(i int) []byte {
		if i%2 == res {
			return ret
		}
		return nil
	}
}
