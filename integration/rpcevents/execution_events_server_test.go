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
	"sync"
	"testing"
	"time"

	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/integration"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutionEventsTest(t *testing.T) {
	kern, shutdown := integration.RunNode(t, rpctest.GenesisDoc, rpctest.PrivateAccounts)
	defer shutdown()
	tcli := rpctest.NewTransactClient(t, kern.GRPCListenAddress().String())
	ecli := rpctest.NewExecutionEventsClient(t, kern.GRPCListenAddress().String())
	inputAddress0 := rpctest.PrivateAccounts[0].GetAddress()
	inputAddress1 := rpctest.PrivateAccounts[1].GetAddress()

	t.Run("Group", func(t *testing.T) {
		t.Run("StreamDB", func(t *testing.T) {
			numSends := 4
			request := &rpcevents.BlocksRequest{
				BlockRange: doSends(t, numSends, tcli, kern, inputAddress1, 2004),
			}
			var blocks []*exec.BlockExecution

			stream, err := ecli.Stream(context.Background(), request)
			require.NoError(t, err)

			err = rpcevents.ConsumeBlockExecutions(stream, func(be *exec.BlockExecution) error {
				blocks = append(blocks, be)
				return nil
			})
			require.Equal(t, io.EOF, err)

			require.True(t, len(blocks) > 0, "should see at least one block")
			var height uint64
			for _, b := range blocks {
				if height > 0 {
					assert.Equal(t, height+1, b.Height)
				}
				for range b.TxExecutions {
					numSends--
				}
				height = b.Height
			}
			require.Equal(t, 0, numSends, "all transactions should be observed")
			require.NoError(t, stream.CloseSend())
		})

		t.Run("Stream_streaming", func(t *testing.T) {
			request := &rpcevents.BlocksRequest{
				BlockRange: rpcevents.NewBlockRange(rpcevents.AbsoluteBound(0), rpcevents.StreamBound()),
			}
			stream, err := ecli.Stream(context.Background(), request)
			require.NoError(t, err)
			batches := 3
			sendsPerBatch := 4
			total := batches * sendsPerBatch
			doneCh := make(chan []struct{})
			go func() {
				for i := 0; i < batches; i++ {
					doSends(t, sendsPerBatch, tcli, kern, inputAddress0, 2004)
				}
				close(doneCh)
			}()

			err = rpcevents.ConsumeBlockExecutions(stream, func(be *exec.BlockExecution) error {
				for range be.TxExecutions {
					total--
				}
				if total == 0 {
					return io.EOF
				}
				return nil
			})
			require.Equal(t, io.EOF, err)
			assert.Equal(t, 0, total)
			require.NoError(t, stream.CloseSend())
			<-doneCh
		})

		t.Run("StreamContains2", func(t *testing.T) {
			request := &rpcevents.BlocksRequest{
				BlockRange: rpcevents.AbsoluteRange(0, 12),
				Query:      "Height CONTAINS '2'",
			}
			stream, err := ecli.Stream(context.Background(), request)
			require.NoError(t, err)
			numSends := 4
			var blocks []*exec.BlockExecution
			require.NoError(t, err)
			doSends(t, numSends, tcli, kern, inputAddress1, 1992)
			require.NoError(t, err)
			err = rpcevents.ConsumeBlockExecutions(stream, func(be *exec.BlockExecution) error {
				blocks = append(blocks, be)
				assert.Contains(t, strconv.FormatUint(be.Height, 10), "2")
				return nil
			})
			require.Equal(t, io.EOF, err)
			require.Len(t, blocks, 2, "should record blocks 2 and 12")
			assert.Equal(t, uint64(2), blocks[0].Height)
			assert.Equal(t, uint64(12), blocks[1].Height)

			require.NoError(t, stream.CloseSend())
		})

		t.Run("GetEventsSend", func(t *testing.T) {
			numSends := 1100
			request := &rpcevents.BlocksRequest{BlockRange: doSends(t, numSends, tcli, kern, inputAddress0, 2004)}
			responses, err := getEvents(t, request, ecli)
			require.NoError(t, err)
			assert.Equal(t, numSends*2, countEventsAndCheckConsecutive(t, responses),
				"should receive 1 input, 1 output per send")
		})

		t.Run("GetEventsSendContainsAA", func(t *testing.T) {
			numSends := 1100
			request := &rpcevents.BlocksRequest{
				BlockRange: doSends(t, numSends, tcli, kern, inputAddress1, 2004),
				Query:      "TxHash CONTAINS 'AA'",
			}
			responses, err := getEvents(t, request, ecli)
			require.NoError(t, err)
			for _, response := range responses {
				for _, ev := range response.Events {
					require.Contains(t, ev.Header.TxHash.String(), "AA")
				}
			}
		})

		t.Run("GetEventsSendFiltered", func(t *testing.T) {
			numSends := 500
			request := &rpcevents.BlocksRequest{
				BlockRange: doSends(t, numSends, tcli, kern, inputAddress0, 999),
				Query: query.NewBuilder().AndEquals("Input.Address", inputAddress0.String()).
					AndEquals(event.EventTypeKey, exec.TypeAccountInput.String()).String(),
			}
			responses, err := getEvents(t, request, ecli)
			require.NoError(t, err)
			assert.Equal(t, numSends, countEventsAndCheckConsecutive(t, responses), "should receive every single input event per send")
		})

		t.Run("Revert", func(t *testing.T) {
			txe, err := rpctest.CreateContract(tcli, inputAddress0, solidity.Bytecode_Revert, nil)
			require.NoError(t, err)
			spec, err := abi.ReadSpec(solidity.Abi_Revert)
			require.NoError(t, err)
			data, _, err := spec.Pack("RevertAt", 4)
			require.NoError(t, err)
			contractAddress := txe.Receipt.ContractAddress
			txe, err = rpctest.CallContract(tcli, inputAddress0, contractAddress, data)
			require.NoError(t, err)
			assert.Equal(t, errors.ErrorCodeExecutionReverted, txe.Exception.Code)
			assert.Contains(t, txe.Exception.Error(), "I have reverted")

			request := &rpcevents.BlocksRequest{
				BlockRange: rpcevents.NewBlockRange(rpcevents.AbsoluteBound(0), rpcevents.LatestBound()),
				Query: query.Must(query.NewBuilder().AndEquals(event.EventIDKey, exec.EventStringLogEvent(contractAddress)).
					AndEquals(event.TxHashKey, txe.TxHash).Query()).String(),
			}
			evs, err := getEvents(t, request, ecli)
			require.NoError(t, err)
			n := countEventsAndCheckConsecutive(t, evs)
			assert.Equal(t, 0, n, "should not see reverted events")
		})
	})
}

func getEvents(t *testing.T, request *rpcevents.BlocksRequest, ecli rpcevents.ExecutionEventsClient) ([]*rpcevents.EventsResponse, error) {
	evs, err := ecli.Events(context.Background(), request)
	require.NoError(t, err)
	var responses []*rpcevents.EventsResponse
	for {
		resp, err := evs.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

func doSends(t *testing.T, numSends int, cli rpctransact.TransactClient, kern *core.Kernel, inputAddress crypto.Address,
	amt uint64) *rpcevents.BlockRange {
	expecter := rpctest.ExpectTxs(kern.Emitter, "doSends")
	wg := new(sync.WaitGroup)
	for i := 0; i < numSends; i++ {
		wg.Add(1)
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
		expecter.Expect(receipt.TxHash)
		assert.False(t, receipt.CreatesContract)
		wg.Done()
	}
	wg.Wait()
	return expecter.AssertCommitted(t)
}

func countEventsAndCheckConsecutive(t *testing.T, responses []*rpcevents.EventsResponse) int {
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
