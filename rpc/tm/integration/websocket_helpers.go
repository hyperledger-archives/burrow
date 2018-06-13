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
	"bytes"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/rpc"
	tmClient "github.com/hyperledger/burrow/rpc/tm/client"
	rpcClient "github.com/hyperledger/burrow/rpc/tm/lib/client"
	rpcTypes "github.com/hyperledger/burrow/rpc/tm/lib/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/require"
	tmTypes "github.com/tendermint/tendermint/types"
)

const (
	timeoutSeconds       = 8
	expectBlockInSeconds = 2
)

//--------------------------------------------------------------------------------
// Utilities for testing the websocket service
type blockPredicate func(block *tmTypes.Block) bool
type resultEventChecker func(eventID string, resultEvent *rpc.ResultEvent) (bool, error)

// create a new connection
func newWSClient() *rpcClient.WSClient {
	wsc := rpcClient.NewWSClient(websocketAddr, websocketEndpoint)
	if err := wsc.Start(); err != nil {
		panic(err)
	}
	return wsc
}

func stopWSClient(wsc *rpcClient.WSClient) {
	wsc.Stop()
}

// subscribe to an event
func subscribe(t *testing.T, wsc *rpcClient.WSClient, eventId string) {
	if err := tmClient.Subscribe(wsc, eventId); err != nil {
		t.Fatal(err)
	}
}

func subscribeAndGetSubscriptionId(t *testing.T, wsc *rpcClient.WSClient, eventId string) string {
	if err := tmClient.Subscribe(wsc, eventId); err != nil {
		t.Fatal(err)
	}

	timeout := time.NewTimer(timeoutSeconds * time.Second)
	for {
		select {
		case <-timeout.C:
			t.Fatal("Timeout waiting for subscription result")
		case response := <-wsc.ResponsesCh:
			require.Nil(t, response.Error, "Got error response from websocket channel: %v", response.Error)
			if response.ID == tmClient.SubscribeRequestID {
				res := new(rpc.ResultSubscribe)
				err := json.Unmarshal(response.Result, res)
				if err == nil {
					return res.SubscriptionID
				}
			}
		}
	}
}

// unsubscribe from an event
func unsubscribe(t *testing.T, wsc *rpcClient.WSClient, subscriptionId string) {
	if err := tmClient.Unsubscribe(wsc, subscriptionId); err != nil {
		t.Fatal(err)
	}
}

// broadcast transaction and wait for new block
func broadcastTxAndWait(t *testing.T, client tmClient.RPCClient, txEnv *txs.Envelope) (*txs.Receipt, error) {
	wsc := newWSClient()
	defer stopWSClient(wsc)
	inputs := txEnv.Tx.GetInputs()
	if len(inputs) == 0 {
		t.Fatalf("cannot broadcastAndWait fot Tx with no inputs")
	}
	address := inputs[0].Address

	var rec *txs.Receipt
	var err error

	err = subscribeAndWaitForNext(t, wsc, events.EventStringAccountInput(address),
		func() {
			rec, err = tmClient.BroadcastTx(client, txEnv)
		}, func(eventID string, resultEvent *rpc.ResultEvent) (bool, error) {
			return true, nil
		})
	if err != nil {
		return nil, err
	}
	return rec, err
}

func nextBlockPredicateFn() blockPredicate {
	initialHeight := int64(-1)
	return func(block *tmTypes.Block) bool {
		if initialHeight <= 0 {
			initialHeight = block.Height
			return false
		} else {
			// TODO: [Silas] remove the + 1 here. It is a workaround for the fact
			// that tendermint fires the NewBlock event before it has finalised its
			// state updates, so we have to wait for the block after the block we
			// want in order for the Tx to be genuinely final.
			// This should be addressed by: https://github.com/tendermint/tendermint/pull/265
			return block.Height > initialHeight+1
		}
	}
}

func waitNBlocks(t *testing.T, wsc *rpcClient.WSClient, n int) {
	i := 0
	require.NoError(t, runThenWaitForBlock(t, wsc,
		func(block *tmTypes.Block) bool {
			i++
			return i >= n
		},
		func() {}))
}

func runThenWaitForBlock(t *testing.T, wsc *rpcClient.WSClient, predicate blockPredicate, runner func()) error {
	eventDataChecker := func(event string, eventData *rpc.ResultEvent) (bool, error) {
		eventDataNewBlock := eventData.Tendermint.EventDataNewBlock()
		if eventDataNewBlock == nil {
			return false, fmt.Errorf("could not convert %#v to EventDataNewBlock", eventData)
		}
		return predicate(eventDataNewBlock.Block), nil
	}
	return subscribeAndWaitForNext(t, wsc, tmTypes.EventNewBlock, runner, eventDataChecker)
}

func subscribeAndWaitForNext(t *testing.T, wsc *rpcClient.WSClient, event string, runner func(),
	checker resultEventChecker) error {

	subId := subscribeAndGetSubscriptionId(t, wsc, event)
	defer unsubscribe(t, wsc, subId)
	return waitForEvent(t, wsc, event, runner, checker)
}

// waitForEvent executes runner that is expected to trigger events. It then
// waits for any events on the supplies WSClient and checks the eventData with
// the eventDataChecker which is a function that is passed the event name
// and the Data and returns the pair of stopWaiting, err. Where if
// stopWaiting is true waitForEvent will return or if stopWaiting is false
// waitForEvent will keep listening for new events. If an error is returned
// waitForEvent will fail the test.
func waitForEvent(t *testing.T, wsc *rpcClient.WSClient, eventID string, runner func(),
	checker resultEventChecker) error {

	// go routine to wait for websocket msg
	eventsCh := make(chan *rpc.ResultEvent)
	shutdownEventsCh := make(chan bool, 1)
	errCh := make(chan error)

	// do stuff (transactions)
	runner()

	// Read message
	go func() {
		for {
			select {
			case <-shutdownEventsCh:
				return
			case r := <-wsc.ResponsesCh:
				if r.Error != nil {
					errCh <- r.Error
					return
				}
				if r.ID == tmClient.EventResponseID(eventID) {
					resultEvent, err := readResponse(r)
					if err != nil {
						errCh <- err
					} else {
						eventsCh <- resultEvent
					}
				}
			case <-wsc.Quit():
				return
			}
		}
	}()

	// Don't block up WSClient
	defer func() { shutdownEventsCh <- true }()

	for {
		select {
		// wait for an event or timeout
		case <-time.After(timeoutSeconds * time.Second):
			t.Fatalf("waitForEvent timed out: %s", debug.Stack())
		case eventData := <-eventsCh:
			// run the check
			stopWaiting, err := checker(eventID, eventData)
			if err != nil {
				return err
			}
			if stopWaiting {
				return nil
			}
		case err := <-errCh:
			return err
		}
	}
	return nil
}

func readResponse(r rpcTypes.RPCResponse) (*rpc.ResultEvent, error) {
	if r.Error != nil {
		return nil, r.Error
	}
	resultEvent := new(rpc.ResultEvent)
	err := json.Unmarshal(r.Result, resultEvent)
	if err != nil {
		return nil, err
	}
	return resultEvent, nil
}

//--------------------------------------------------------------------------------

func unmarshalValidateSend(amt uint64, toAddr crypto.Address, resultEvent *rpc.ResultEvent) error {
	data := resultEvent.EventDataTx
	if data == nil {
		return fmt.Errorf("event data %v is not EventDataTx", resultEvent)
	}
	if data.Exception != "" {
		return fmt.Errorf(data.Exception)
	}
	tx := data.Tx.Payload.(*payload.SendTx)
	if tx.Inputs[0].Address != privateAccounts[0].Address() {
		return fmt.Errorf("senders do not match up! Got %s, expected %s", tx.Inputs[0].Address,
			privateAccounts[0].Address())
	}
	if tx.Inputs[0].Amount != amt {
		return fmt.Errorf("amt does not match up! Got %d, expected %d", tx.Inputs[0].Amount, amt)
	}
	if tx.Outputs[0].Address != toAddr {
		return fmt.Errorf("receivers do not match up! Got %s, expected %s", tx.Outputs[0].Address,
			privateAccounts[0].Address())
	}
	return nil
}

func unmarshalValidateTx(amt uint64, returnCode []byte) resultEventChecker {
	return func(eventID string, resultEvent *rpc.ResultEvent) (bool, error) {
		data := resultEvent.EventDataTx
		if data == nil {
			return true, fmt.Errorf("event data %v is not EventDataTx", *resultEvent)
		}
		if data.Exception != "" {
			return true, fmt.Errorf(data.Exception)
		}
		tx := data.Tx.Payload.(*payload.CallTx)
		if tx.Input.Address != privateAccounts[0].Address() {
			return true, fmt.Errorf("senders do not match up! Got %s, expected %s",
				tx.Input.Address, privateAccounts[0].Address())
		}
		if tx.Input.Amount != amt {
			return true, fmt.Errorf("amount does not match up! Got %d, expected %d",
				tx.Input.Amount, amt)
		}
		ret := data.Return
		if !bytes.Equal(ret, returnCode) {
			return true, fmt.Errorf("tx did not return correctly. Got %x, expected %x", ret, returnCode)
		}
		return true, nil
	}
}

func unmarshalValidateCall(origin crypto.Address, returnCode []byte, txid *[]byte) resultEventChecker {
	return func(eventID string, resultEvent *rpc.ResultEvent) (bool, error) {
		data := resultEvent.EventDataCall
		if data == nil {
			return true, fmt.Errorf("event data %v is not EventDataTx", *resultEvent)
		}
		if data.Exception != "" {
			return true, fmt.Errorf(data.Exception)
		}
		if data.Origin != origin {
			return true, fmt.Errorf("origin does not match up! Got %s, expected %s", data.Origin, origin)
		}
		ret := data.Return
		if !bytes.Equal(ret, returnCode) {
			return true, fmt.Errorf("call did not return correctly. Got %x, expected %x", ret, returnCode)
		}
		if !bytes.Equal(data.TxHash, *txid) {
			return true, fmt.Errorf("TxIDs do not match up! Got %x, expected %x",
				data.TxHash, *txid)
		}
		return true, nil
	}
}
