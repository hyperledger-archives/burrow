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

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/rpc/lib/client"
	tm_types "github.com/tendermint/tendermint/types"
)

const (
	timeoutSeconds       = 2
	expectBlockInSeconds = timeoutSeconds
)

//--------------------------------------------------------------------------------
// Utilities for testing the websocket service
type blockPredicate func(block *tm_types.Block) bool

// create a new connection
func newWSClient() *rpcclient.WSClient {
	wsc := rpcclient.NewWSClient(websocketAddr, websocketEndpoint)
	if _, err := wsc.Start(); err != nil {
		panic(err)
	}
	return wsc
}

// subscribe to an event
func subscribe(t *testing.T, wsc *rpcclient.WSClient, eventId string) {
	if err := Subscribe(wsc, eventId); err != nil {
		t.Fatal(err)
	}
}

func subscribeAndGetSubscriptionId(t *testing.T, wsc *rpcclient.WSClient,
	eventId string) string {
	if err := Subscribe(wsc, eventId); err != nil {
		t.Fatal(err)
	}

	timeout := time.NewTimer(timeoutSeconds * time.Second)
	for {
		select {
		case <-timeout.C:
			t.Fatal("Timeout waiting for subscription result")
		case bs := <-wsc.ResultsCh:
			res := new(rpc.ResultSubscribe)
			readResult(t, bs, res)
			return res.SubscriptionId
		}
	}
}

// unsubscribe from an event
func unsubscribe(t *testing.T, wsc *rpcclient.WSClient, subscriptionId string) {
	if err := Unsubscribe(wsc, subscriptionId); err != nil {
		t.Fatal(err)
	}
}

// broadcast transaction and wait for new block
func broadcastTxAndWaitForBlock(t *testing.T, client RPCClient, wsc *rpcclient.WSClient,
	tx txs.Tx) (*txs.Receipt, error) {

	var rec *txs.Receipt
	var err error
	runThenWaitForBlock(t, wsc, nextBlockPredicateFn(),
		func() {
			rec, err = BroadcastTx(client, tx)
		})
	return rec, err
}

func nextBlockPredicateFn() blockPredicate {
	initialHeight := -1
	return func(block *tm_types.Block) bool {
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

func waitNBlocks(t *testing.T, wsc *rpcclient.WSClient, n int) {
	i := 0
	runThenWaitForBlock(t, wsc,
		func(block *tm_types.Block) bool {
			i++
			return i >= n
		},
		func() {})
}

func runThenWaitForBlock(t *testing.T, wsc *rpcclient.WSClient,
	predicate blockPredicate, runner func()) {
	subscribeAndWaitForNext(t, wsc, tm_types.EventStringNewBlock(),
		runner,
		func(event string, eventData event.AnyEventData) (bool, error) {
			eventDataNewBlock := eventData.EventDataNewBlock()
			if eventDataNewBlock == nil {
				return false, fmt.Errorf("could not convert %s to EventDataNewBlock", eventData)
			}
			return predicate(eventDataNewBlock.Block), nil
		})
}

func subscribeAndWaitForNext(t *testing.T, wsc *rpcclient.WSClient, event string,
	runner func(),
	eventDataChecker func(string, event.AnyEventData) (bool, error)) {
	subId := subscribeAndGetSubscriptionId(t, wsc, event)
	defer unsubscribe(t, wsc, subId)
	waitForEvent(t,
		wsc,
		event,
		runner,
		eventDataChecker)
}

// waitForEvent executes runner that is expected to trigger events. It then
// waits for any events on the supplies WSClient and checks the eventData with
// the eventDataChecker which is a function that is passed the event name
// and the Data and returns the pair of stopWaiting, err. Where if
// stopWaiting is true waitForEvent will return or if stopWaiting is false
// waitForEvent will keep listening for new events. If an error is returned
// waitForEvent will fail the test.
func waitForEvent(t *testing.T, wsc *rpcclient.WSClient, eventid string,
	runner func(),
	eventDataChecker func(string, event.AnyEventData) (bool, error)) waitForEventResult {

	// go routine to wait for websocket msg
	eventsCh := make(chan event.AnyEventData)
	shutdownEventsCh := make(chan bool, 1)
	errCh := make(chan error)

	// do stuff (transactions)
	runner()

	// Read message
	go func() {
	LOOP:
		for {
			select {
			case <-shutdownEventsCh:
				break LOOP
			case r := <-wsc.ResultsCh:
				resultEvent := new(rpc.ResultEvent)
				readResult(t, r, resultEvent)
				if resultEvent != nil && resultEvent.Event == eventid {
					// Keep feeding events
					eventsCh <- resultEvent.Data
				}
			case err := <-wsc.ErrorsCh:
				errCh <- err
				break LOOP
			case <-wsc.Quit:
				break LOOP
			}
		}
	}()

	// Don't block up WSClient
	defer func() { shutdownEventsCh <- true }()

	for {
		select {
		// wait for an event or timeout
		case <-time.After(timeoutSeconds * time.Second):
			return waitForEventResult{timeout: true}
		case eventData := <-eventsCh:
			// run the check
			stopWaiting, err := eventDataChecker(eventid, eventData)
			if err != nil {
				t.Fatal(err) // Show the stack trace.
			}
			if stopWaiting {
				return waitForEventResult{}
			}
		case err := <-errCh:
			t.Fatal(err)
		}
	}
}

type waitForEventResult struct {
	error
	timeout bool
}

func (err waitForEventResult) Timeout() bool {
	return err.timeout
}

//--------------------------------------------------------------------------------

func unmarshalValidateSend(amt uint64, toAddr acm.Address) func(string, event.AnyEventData) (bool, error) {
	return func(eid string, eventData event.AnyEventData) (bool, error) {
		data := eventData.EventDataTx()
		if data == nil {
			return true, fmt.Errorf("event data %s is not EventDataTx", eventData)
		}
		if data.Exception != "" {
			return true, fmt.Errorf(data.Exception)
		}
		tx := data.Tx.(*txs.SendTx)
		if tx.Inputs[0].Address != privateAccounts[0].Address() {
			return true, fmt.Errorf("senders do not match up! Got %s, expected %s", tx.Inputs[0].Address,
				privateAccounts[0].Address)
		}
		if tx.Inputs[0].Amount != amt {
			return true, fmt.Errorf("amt does not match up! Got %d, expected %d", tx.Inputs[0].Amount, amt)
		}
		if tx.Outputs[0].Address != toAddr {
			return true, fmt.Errorf("receivers do not match up! Got %s, expected %s", tx.Outputs[0].Address,
				privateAccounts[0].Address)
		}
		return true, nil
	}
}

func unmarshalValidateTx(amt uint64, returnCode []byte) func(string, event.AnyEventData) (bool, error) {
	return func(eid string, eventData event.AnyEventData) (bool, error) {
		data := eventData.EventDataTx()
		if data == nil {
			return true, fmt.Errorf("event data %s is not EventDataTx", eventData)
		}
		if data.Exception != "" {
			return true, fmt.Errorf(data.Exception)
		}
		tx := data.Tx.(*txs.CallTx)
		if tx.Input.Address != privateAccounts[0].Address() {
			return true, fmt.Errorf("Senders do not match up! Got %x, expected %x",
				tx.Input.Address, privateAccounts[0].Address)
		}
		if tx.Input.Amount != amt {
			return true, fmt.Errorf("Amt does not match up! Got %d, expected %d",
				tx.Input.Amount, amt)
		}
		ret := data.Return
		if !bytes.Equal(ret, returnCode) {
			return true, fmt.Errorf("Tx did not return correctly. Got %x, expected %x", ret, returnCode)
		}
		return true, nil
	}
}

func unmarshalValidateCall(origin acm.Address, returnCode []byte, txid *[]byte) func(string, event.AnyEventData) (bool, error) {
	return func(eid string, eventData event.AnyEventData) (bool, error) {
		data := eventData.EventDataCall()
		if data == nil {
			return true, fmt.Errorf("event data %s is not EventDataTx", eventData)
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
		if !bytes.Equal(data.TxID, *txid) {
			return true, fmt.Errorf("TxIDs do not match up! Got %x, expected %x",
				data.TxID, *txid)
		}
		return true, nil
	}
}

func readResult(t *testing.T, bs []byte, result interface{}) {
	err := json.Unmarshal(bs, result)
	if err != nil {
		t.Fatal(err)
	}
}
