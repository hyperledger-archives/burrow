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
	"testing"

	"io"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/events/pbevents"
	"github.com/hyperledger/burrow/execution/pbtransactor"
	"github.com/hyperledger/burrow/rpc/test"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutionEventsSendStream(t *testing.T) {
	request := &pbevents.GetEventsRequest{
		BlockRange: pbevents.NewBlockRange(pbevents.LatestBound(), pbevents.StreamBound()),
	}
	tcli := test.NewTransactorClient(t)
	ecli := test.NewExecutionEventsClient(t)
	stream, err := ecli.GetEvents(context.Background(), request)
	require.NoError(t, err)
	numSends := 1
	for i := 0; i < numSends; i++ {
		doSends(t, 2, tcli)
		response, err := stream.Recv()
		require.NoError(t, err)
		require.Len(t, response.Events, 4, "expect multiple events")
		assert.Equal(t, payload.TypeSend.String(), response.Events[0].GetHeader().GetTxType())
		assert.Equal(t, payload.TypeSend.String(), response.Events[1].GetHeader().GetTxType())
	}
	require.NoError(t, stream.CloseSend())
}

func TestExecutionEventsSend(t *testing.T) {
	request := &pbevents.GetEventsRequest{
		BlockRange: pbevents.NewBlockRange(pbevents.AbsoluteBound(kern.Blockchain.LastBlockHeight()),
			pbevents.AbsoluteBound(300)),
	}
	numSends := 1100
	responses := testExecutionEventsSend(t, numSends, request)
	assert.Equal(t, numSends*2, totalEvents(responses), "should receive and input and output event per send")
}

func TestExecutionEventsSendFiltered(t *testing.T) {
	request := &pbevents.GetEventsRequest{
		BlockRange: pbevents.NewBlockRange(pbevents.AbsoluteBound(kern.Blockchain.LastBlockHeight()),
			pbevents.AbsoluteBound(300)),
		Query: query.NewBuilder().AndEquals(event.EventTypeKey, events.TypeAccountInput.String()).String(),
	}
	numSends := 500
	responses := testExecutionEventsSend(t, numSends, request)
	assert.Equal(t, numSends, totalEvents(responses), "should receive a single input event per send")
}

func testExecutionEventsSend(t *testing.T, numSends int, request *pbevents.GetEventsRequest) []*pbevents.GetEventsResponse {
	doSends(t, numSends, test.NewTransactorClient(t))
	ecli := test.NewExecutionEventsClient(t)

	evs, err := ecli.GetEvents(context.Background(), request)
	require.NoError(t, err)
	var responses []*pbevents.GetEventsResponse
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

func doSends(t *testing.T, numSends int, cli pbtransactor.TransactorClient) {
	countCh := test.CommittedTxCount(t, kern.Emitter)
	for i := 0; i < numSends; i++ {
		send, err := cli.Send(context.Background(), &pbtransactor.SendParam{
			InputAccount: inputAccount,
			Amount:       2003,
			ToAddress:    privateAccounts[3].Address().Bytes(),
		})
		require.NoError(t, err)
		assert.Equal(t, false, send.CreatesContract)
	}
	require.Equal(t, numSends, <-countCh)
}

func totalEvents(respones []*pbevents.GetEventsResponse) int {
	i := 0
	for _, resp := range respones {
		i += len(resp.Events)
	}
	return i
}
