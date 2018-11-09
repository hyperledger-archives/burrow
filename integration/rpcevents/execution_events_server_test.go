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

package rpcevents

import (
	"context"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBlocks(t *testing.T) {
	request := &rpcevents.BlocksRequest{
		BlockRange: rpcevents.NewBlockRange(rpcevents.LatestBound(), rpcevents.StreamBound()),
	}
	tcli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	ecli := rpctest.NewExecutionEventsClient(t, testConfig.RPC.GRPC.ListenAddress)
	stream, err := ecli.GetBlocks(context.Background(), request)
	require.NoError(t, err)
	pulls := 3
	sendsPerPull := 4
	var blocks []*exec.BlockExecution
	for i := 0; i < pulls; i++ {
		doSends(t, sendsPerPull, tcli)
		block, err := stream.Recv()
		require.NoError(t, err)
		blocks = append(blocks, block)
	}
	assert.True(t, len(blocks) > 0, "should see at least one block (height 2)")
	var height uint64
	for _, b := range blocks {
		if height > 0 {
			assert.Equal(t, height+1, b.Height)
		}
		height = b.Height
	}
	require.NoError(t, stream.CloseSend())
}
func TestGetBlocksContains2(t *testing.T) {
	request := &rpcevents.BlocksRequest{
		BlockRange: rpcevents.NewBlockRange(rpcevents.LatestBound(), rpcevents.StreamBound()),
		Query:      "Height CONTAINS '2'",
	}
	tcli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	ecli := rpctest.NewExecutionEventsClient(t, testConfig.RPC.GRPC.ListenAddress)
	stream, err := ecli.GetBlocks(context.Background(), request)
	require.NoError(t, err)
	pulls := 2
	sendsPerPull := 4
	var blocks []*exec.BlockExecution
	for i := 0; i < pulls; i++ {
		doSends(t, sendsPerPull, tcli)
		block, err := stream.Recv()
		require.NoError(t, err)
		blocks = append(blocks, block)
		assert.Contains(t, strconv.FormatUint(block.Height, 10), "2")
	}
	assert.True(t, len(blocks) > 0, "should see at least one block (height 2)")
	require.NoError(t, stream.CloseSend())
}

func TestGetEventsSend(t *testing.T) {
	request := &rpcevents.BlocksRequest{
		BlockRange: rpcevents.NewBlockRange(rpcevents.AbsoluteBound(kern.Blockchain.LastBlockHeight()),
			rpcevents.LatestBound()),
	}
	numSends := 1100
	doSends(t, numSends, rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress))
	responses := getEvents(t, request)
	assert.Equal(t, numSends*2, countEventsAndCheckConsecutive(t, responses),
		"should receive 1 input, 1 output per send")
}

func TestGetEventsSendContainsAA(t *testing.T) {
	request := &rpcevents.BlocksRequest{
		BlockRange: rpcevents.NewBlockRange(rpcevents.AbsoluteBound(kern.Blockchain.LastBlockHeight()),
			rpcevents.LatestBound()),
		Query: "TxHash CONTAINS 'AA'",
	}
	numSends := 1100
	doSends(t, numSends, rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress))
	responses := getEvents(t, request)
	for _, response := range responses {
		for _, ev := range response.Events {
			require.Contains(t, ev.Header.TxHash.String(), "AA")
		}
	}
}

func TestGetEventsSendFiltered(t *testing.T) {
	request := &rpcevents.BlocksRequest{
		BlockRange: rpcevents.NewBlockRange(rpcevents.AbsoluteBound(kern.Blockchain.LastBlockHeight()),
			rpcevents.LatestBound()),
		Query: query.NewBuilder().AndEquals(event.EventTypeKey, exec.TypeAccountInput.String()).String(),
	}
	numSends := 500
	doSends(t, numSends, rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress))
	responses := getEvents(t, request)
	assert.Equal(t, numSends, countEventsAndCheckConsecutive(t, responses), "should receive a single input event per send")
}

func TestRevert(t *testing.T) {
	tcli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	txe := rpctest.CreateContract(t, tcli, inputAddress, solidity.Bytecode_Revert)
	spec, err := abi.ReadAbiSpec(solidity.Abi_Revert)
	require.NoError(t, err)
	data, err := spec.Pack("RevertAt", 4)
	require.NoError(t, err)
	contractAddress := txe.Receipt.ContractAddress
	txe = rpctest.CallContract(t, tcli, inputAddress, contractAddress, data)
	assert.Equal(t, errors.ErrorCodeExecutionReverted, txe.Exception.Code)

	request := &rpcevents.BlocksRequest{
		BlockRange: rpcevents.NewBlockRange(rpcevents.AbsoluteBound(0), rpcevents.LatestBound()),
		Query: query.Must(query.NewBuilder().AndEquals(event.EventIDKey, exec.EventStringLogEvent(contractAddress)).
			AndEquals(event.TxHashKey, txe.TxHash).Query()).String(),
	}
	evs := getEvents(t, request)
	n := countEventsAndCheckConsecutive(t, evs)
	assert.Equal(t, 0, n, "should not see reverted events")
}

func getEvents(t *testing.T, request *rpcevents.BlocksRequest) []*rpcevents.GetEventsResponse {
	ecli := rpctest.NewExecutionEventsClient(t, testConfig.RPC.GRPC.ListenAddress)
	evs, err := ecli.GetEvents(context.Background(), request)
	require.NoError(t, err)
	var responses []*rpcevents.GetEventsResponse
	for {
		resp, err := evs.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}
		responses = append(responses, resp)
	}
	return responses
}

func doSends(t *testing.T, numSends int, cli rpctransact.TransactClient) {
	countCh := rpctest.CommittedTxCount(t, kern.Emitter)
	amt := uint64(2004)
	for i := 0; i < numSends; i++ {
		// Slow us down a bit to ensure spread across blocks
		time.Sleep(time.Millisecond)
		receipt, err := cli.SendTxAsync(context.Background(),
			&payload.SendTx{
				Inputs: []*payload.TxInput{{
					Address: inputAddress,
					Amount:  amt,
				}},
				Outputs: []*payload.TxOutput{{
					Address: rpctest.PrivateAccounts[4].GetAddress(),
					Amount:  amt,
				}},
			})
		require.NoError(t, err)
		assert.False(t, receipt.CreatesContract)
	}
	require.Equal(t, numSends, <-countCh)
}

func countEventsAndCheckConsecutive(t *testing.T, responses []*rpcevents.GetEventsResponse) int {
	i := 0
	var height uint64
	var index uint64
	txHash := ""
	for _, resp := range responses {
		require.True(t, resp.Height > height, "must not receive multiple GetEventsResponses for the same block")
		height := resp.Height
		for _, ev := range resp.Events {
			require.Equal(t, ev.Header.Height, height, "height of event headers much match height of GetEventsResponse")
			if txHash != ev.Header.TxHash.String() {
				txHash = ev.Header.TxHash.String()
				index = 0
			}
			if ev.Header.Index > index {
				require.Equal(t, index+1, ev.Header.Index)
				index++
			}
			i++
		}
	}
	return i
}
