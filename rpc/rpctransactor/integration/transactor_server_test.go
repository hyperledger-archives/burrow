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
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/pbtransactor"
	"github.com/hyperledger/burrow/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
)

func TestTransactCallNoCode(t *testing.T) {
	cli := newClient(t)

	// Flip flops between sending private key and input address to test private key and address based signing
	toAddress := privateAccounts[2].Address()

	numCreates := 1000
	countCh := committedTxCount(t)
	for i := 0; i < numCreates; i++ {
		receipt, err := cli.Transact(context.Background(), &pbtransactor.TransactParam{
			InputAccount: inputAccount(i),
			Address:      toAddress.Bytes(),
			Data:         []byte{},
			Fee:          2,
			GasLimit:     10000 + uint64(i),
		})
		require.NoError(t, err)
		assert.False(t, receipt.CreatesContract)
		assert.Equal(t, toAddress.Bytes(), receipt.ContractAddress)
	}
	require.Equal(t, numCreates, <-countCh)
}

func TestTransactCreate(t *testing.T) {
	cli := newClient(t)
	numGoroutines := 100
	numCreates := 50
	wg := new(sync.WaitGroup)
	wg.Add(numGoroutines)
	// Flip flops between sending private key and input address to test private key and address based signing
	bc, err := hex.DecodeString(strangeLoopBytecode)
	require.NoError(t, err)
	countCh := committedTxCount(t)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numCreates; j++ {
				create, err := cli.Transact(context.Background(), &pbtransactor.TransactParam{
					InputAccount: inputAccount(i),
					Address:      nil,
					Data:         bc,
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
	cli := newClient(b)
	bc, err := hex.DecodeString(strangeLoopBytecode)
	require.NoError(b, err)
	for i := 0; i < b.N; i++ {
		create, err := cli.Transact(context.Background(), &pbtransactor.TransactParam{
			InputAccount: inputAccount(i),
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
	cli := newClient(t)
	bc, err := hex.DecodeString(strangeLoopBytecode)
	require.NoError(t, err)
	numGoroutines := 5
	numRuns := 2
	countCh := committedTxCount(t)
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numRuns; j++ {
			create, err := cli.TransactAndHold(context.Background(), &pbtransactor.TransactParam{
				InputAccount: inputAccount(i),
				Address:      nil,
				Data:         bc,
				Fee:          2,
				GasLimit:     10000,
			})
			require.NoError(t, err)
			assert.Equal(t, uint64(0), create.StackDepth)
			functionID := abi.FunctionID("UpsieDownsie()")
			call, err := cli.TransactAndHold(context.Background(), &pbtransactor.TransactParam{
				InputAccount: inputAccount(i),
				Address:      create.CallData.Callee,
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
	cli := newClient(t)
	numSends := 1000
	countCh := committedTxCount(t)
	for i := 0; i < numSends; i++ {
		send, err := cli.Send(context.Background(), &pbtransactor.SendParam{
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
	cli := newClient(t)
	for i := 0; i < 2; i++ {
		send, err := cli.SendAndHold(context.Background(), &pbtransactor.SendParam{
			InputAccount: inputAccount(i),
			Amount:       2003,
			ToAddress:    privateAccounts[3].Address().Bytes(),
		})
		require.NoError(t, err)
		assert.Equal(t, false, send.CreatesContract)
	}
}

// Helpers
func newClient(t testing.TB) pbtransactor.TransactorClient {
	conn, err := grpc.Dial(rpc.DefaultGRPCConfig().ListenAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return pbtransactor.NewTransactorClient(conn)
}

var committedTxCountIndex = 0

func committedTxCount(t *testing.T) chan int {
	var numTxs int64
	emptyBlocks := 0
	maxEmptyBlocks := 2
	outCh := make(chan int)
	ch := make(chan *types.EventDataNewBlock)
	ctx := context.Background()
	subscriber := fmt.Sprintf("committedTxCount_%v", committedTxCountIndex)
	committedTxCountIndex++
	require.NoError(t, tendermint.SubscribeNewBlock(ctx, kern.Emitter, subscriber, ch))

	go func() {
		for ed := range ch {
			if ed.Block.NumTxs == 0 {
				emptyBlocks++
			} else {
				emptyBlocks = 0
			}
			if emptyBlocks > maxEmptyBlocks {
				break
			}
			numTxs += ed.Block.NumTxs
			t.Logf("Total TXs committed at block %v: %v (+%v)\n", ed.Block.Height, numTxs, ed.Block.NumTxs)
		}
		require.NoError(t, kern.Emitter.UnsubscribeAll(ctx, subscriber))
		outCh <- int(numTxs)
	}()
	return outCh
}

var inputPrivateKey = privateAccounts[0].PrivateKey().RawBytes()
var inputAddress = privateAccounts[0].Address().Bytes()

func inputAccount(i int) *pbtransactor.InputAccount {
	ia := new(pbtransactor.InputAccount)
	if i%2 == 0 {
		ia.PrivateKey = inputPrivateKey
	} else {
		ia.Address = inputAddress
	}
	return ia
}
