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
	"time"

	"encoding/json"

	"strings"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/events/pbevents"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmTypes "github.com/tendermint/tendermint/types"
)

const eventWait = 2 * time.Second

func TestEventSubscribeNewBlock(t *testing.T) {
	testEventSubscribe(t, tmTypes.EventNewBlock, nil, func(evs []*pbevents.Event) {
		require.True(t, len(evs) > 0, "event poll should return at least 1 event")
		tendermintEvent := new(rpc.TendermintEvent)
		tendermintEventJSON := evs[0].GetTendermintEventJSON()
		require.NoError(t, json.Unmarshal([]byte(tendermintEventJSON), tendermintEvent))
		newBlock, ok := tendermintEvent.TMEventData.(tmTypes.EventDataNewBlock)
		require.True(t, ok, "new block event expected")
		assert.Equal(t, genesisDoc.ChainID(), newBlock.Block.ChainID)
	})
}

func TestEventSubscribeCall(t *testing.T) {
	cli := test.NewTransactorClient(t)
	create := test.CreateContract(t, cli, inputAccount)
	address, err := crypto.AddressFromBytes(create.CallData.Callee)
	require.NoError(t, err)
	testEventSubscribe(t, events.EventStringAccountCall(address),
		func() {
			t.Logf("Calling contract at: %v", address)
			test.CallContract(t, cli, inputAccount, address.Bytes())
		},
		func(evs []*pbevents.Event) {
			require.Len(t, evs, test.UpsieDownsieCallCount, "should see 30 recursive call events")
			for i, ev := range evs {
				assert.Equal(t, uint64(test.UpsieDownsieCallCount-i-1), ev.GetExecutionEvent().GetEventDataCall().GetStackDepth())
			}
		})
}

func TestEventSubscribeLog(t *testing.T) {
	cli := test.NewTransactorClient(t)
	create := test.CreateContract(t, cli, inputAccount)
	address, err := crypto.AddressFromBytes(create.CallData.Callee)
	require.NoError(t, err)
	testEventSubscribe(t, events.EventStringLogEvent(address),
		func() {
			t.Logf("Calling contract at: %v", address)
			test.CallContract(t, cli, inputAccount, address.Bytes())
		},
		func(evs []*pbevents.Event) {
			require.Len(t, evs, test.UpsieDownsieCallCount-2)
			log := evs[0].GetExecutionEvent().GetEventDataLog()
			depth := binary.Int64FromWord256(binary.LeftPadWord256(log.Topics[2]))
			direction := strings.TrimRight(string(log.Topics[1]), "\x00")
			assert.Equal(t, int64(18), depth)
			assert.Equal(t, "Upsie!", direction)
		})
}

func testEventSubscribe(t *testing.T, eventID string, runner func(), eventsCallback func(evs []*pbevents.Event)) {
	cli := test.NewEventsClient(t)
	t.Logf("Subscribing to event: %s", eventID)
	sub, err := cli.EventSubscribe(context.Background(), &pbevents.EventIdParam{
		EventId: eventID,
	})
	require.NoError(t, err)
	defer cli.EventUnsubscribe(context.Background(), sub)

	if runner != nil {
		go runner()
	}

	pollCh := make(chan *pbevents.PollResponse)
	errCh := make(chan error)
	// Catch a single block of events
	go func() {
		for {
			time.Sleep(eventWait)
			poll, err := cli.EventPoll(context.Background(), sub)
			if err != nil {
				errCh <- err
				return
			}
			if len(poll.Events) > 0 {
				pollCh <- poll
			}
		}
	}()
	//var evs []*pbevents.Event
	select {
	case err := <-errCh:
		require.NoError(t, err, "should be no error from EventPoll")
	case poll := <-pollCh:
		eventsCallback(poll.Events)
	case <-time.After(2 * eventWait):
		t.Fatal("timed out waiting for poll event")
	}
}
