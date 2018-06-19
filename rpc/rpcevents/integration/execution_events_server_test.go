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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var committedTxCountIndex = 0
var inputAccount = &pbtransactor.InputAccount{Address: privateAccounts[0].Address().Bytes()}

func TestSend(t *testing.T) {
	tcli := test.NewTransactorClient(t)
	numSends := 1500
	countCh := test.CommittedTxCount(t, kern.Emitter, &committedTxCountIndex)
	for i := 0; i < numSends; i++ {
		send, err := tcli.Send(context.Background(), &pbtransactor.SendParam{
			InputAccount: inputAccount,
			Amount:       2003,
			ToAddress:    privateAccounts[3].Address().Bytes(),
		})
		require.NoError(t, err)
		assert.Equal(t, false, send.CreatesContract)
	}
	require.Equal(t, numSends, <-countCh)

	ecli := test.NewExecutionEventsClient(t)

	evs, err := ecli.GetEvents(context.Background(), &pbevents.GetEventsRequest{
		BlockRange: pbevents.SimpleBlockRange(0, 100),
	})
	require.NoError(t, err)
	i := 0
	for {
		resp, err := evs.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}
		for _, ev := range resp.Events {
			assert.Len(t, ev.Header.TxHash, 20)
			i++
		}
	}
	// Input/output events for each
	assert.Equal(t, numSends*2, i)
}

func TestSendFiltered(t *testing.T) {
	tcli := test.NewTransactorClient(t)
	numSends := 1500
	countCh := test.CommittedTxCount(t, kern.Emitter, &committedTxCountIndex)
	for i := 0; i < numSends; i++ {
		send, err := tcli.Send(context.Background(), &pbtransactor.SendParam{
			InputAccount: inputAccount,
			Amount:       2003,
			ToAddress:    privateAccounts[3].Address().Bytes(),
		})
		require.NoError(t, err)
		assert.Equal(t, false, send.CreatesContract)
	}
	require.Equal(t, numSends, <-countCh)

	ecli := test.NewExecutionEventsClient(t)

	evs, err := ecli.GetEvents(context.Background(), &pbevents.GetEventsRequest{
		BlockRange: pbevents.SimpleBlockRange(0, 100),
		Query:      query.NewBuilder().AndEquals(event.EventTypeKey, events.TypeAccountInput.String()).String(),
	})
	require.NoError(t, err)
	i := 0
	for {
		resp, err := evs.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}
		for _, ev := range resp.Events {
			assert.Len(t, ev.Header.TxHash, 20)
			i++
		}
	}
	// Should only get input events
	assert.Equal(t, numSends, i)
}
