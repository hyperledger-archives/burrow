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
	"encoding/json"
	"testing"
	"time"

	"github.com/hyperledger/burrow/execution/events/pbevents"
	"github.com/hyperledger/burrow/execution/pbtransactor"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmTypes "github.com/tendermint/tendermint/types"
	"github.com/tmthrgd/go-hex"
)

func TestEventSubscribe(t *testing.T) {
	cli := test.NewEventsClient(t)
	sub, err := cli.EventSubscribe(context.Background(), &pbevents.EventIdParam{
		EventId: tmTypes.EventNewBlock,
	})
	require.NoError(t, err)
	defer cli.EventUnsubscribe(context.Background(), sub)

	pollCh := make(chan *pbevents.PollResponse)
	go func() {
		poll := new(pbevents.PollResponse)
		for len(poll.Events) == 0 {
			poll, err = cli.EventPoll(context.Background(), sub)
			require.NoError(t, err)
			time.Sleep(1)
		}
		pollCh <- poll
	}()
	select {
	case poll := <-pollCh:
		require.True(t, len(poll.Events) > 0, "event poll should return at least 1 event")
		tendermintEvent := new(rpc.TendermintEvent)
		tendermintEventJSON := poll.Events[0].GetTendermintEventJSON()
		require.NoError(t, json.Unmarshal([]byte(tendermintEventJSON), tendermintEvent))
		newBlock, ok := tendermintEvent.TMEventData.(tmTypes.EventDataNewBlock)
		require.True(t, ok, "new block event expected")
		assert.Equal(t, genesisDoc.ChainID(), newBlock.Block.ChainID)
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for poll event")
	}
}

func testEventsCall(t *testing.T, numTxs int) {
	cli := test.NewTransactorClient(t)

	bc, err := hex.DecodeString(test.StrangeLoopByteCode)

	require.NoError(t, err)

	countCh := test.CommittedTxCount(t, kern.Emitter)
	for i := 0; i < numTxs; i++ {
		_, err := cli.Transact(context.Background(), &pbtransactor.TransactParam{
			InputAccount: inputAccount,
			Address:      nil,
			Data:         bc,
			Fee:          2,
			GasLimit:     10000,
		})
		require.NoError(t, err)
	}
	require.Equal(t, numTxs, <-countCh)
}
