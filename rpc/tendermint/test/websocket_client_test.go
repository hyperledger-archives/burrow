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

package test

import (
	"fmt"
	"testing"

	"time"

	core_types "github.com/hyperledger/burrow/rpc/tendermint/core/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	_ "github.com/tendermint/tendermint/config/tendermint_test"
)

//--------------------------------------------------------------------------------
// Test the websocket service

// make a simple connection to the server
func TestWSConnect(t *testing.T) {
	wsc := newWSClient()
	wsc.Stop()
}

// receive a new block message
func TestWSNewBlock(t *testing.T) {
	wsc := newWSClient()
	eid := txs.EventStringNewBlock()
	subId := subscribeAndGetSubscriptionId(t, wsc, eid)
	defer func() {
		unsubscribe(t, wsc, subId)
		wsc.Stop()
	}()
	waitForEvent(t, wsc, eid, func() {},
		func(eid string, eventData txs.EventData) (bool, error) {
			fmt.Println("Check: ", eventData.(txs.EventDataNewBlock).Block)
			return true, nil
		})
}

// receive a few new block messages in a row, with increasing height
func TestWSBlockchainGrowth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	eid := txs.EventStringNewBlock()
	subId := subscribeAndGetSubscriptionId(t, wsc, eid)
	defer func() {
		unsubscribe(t, wsc, subId)
		wsc.Stop()
	}()
	// listen for NewBlock, ensure height increases by 1
	var initBlockN int
	for i := 0; i < 2; i++ {
		waitForEvent(t, wsc, eid, func() {},
			func(eid string, eventData txs.EventData) (bool, error) {
				eventDataNewBlock, ok := eventData.(txs.EventDataNewBlock)
				if !ok {
					t.Fatalf("Was expecting EventDataNewBlock but got %v", eventData)
				}
				block := eventDataNewBlock.Block
				if i == 0 {
					initBlockN = block.Height
				} else {
					if block.Header.Height != initBlockN+i {
						return true, fmt.Errorf("Expected block %d, got block %d", i,
							block.Header.Height)
					}
				}

				return true, nil
			})
	}
}

// send a transaction and validate the events from listening for both sender and receiver
func TestWSSend(t *testing.T) {
	wsc := newWSClient()
	toAddr := user[1].Address
	amt := int64(100)
	eidInput := txs.EventStringAccInput(user[0].Address)
	eidOutput := txs.EventStringAccOutput(toAddr)
	subIdInput := subscribeAndGetSubscriptionId(t, wsc, eidInput)
	subIdOutput := subscribeAndGetSubscriptionId(t, wsc, eidOutput)
	defer func() {
		unsubscribe(t, wsc, subIdInput)
		unsubscribe(t, wsc, subIdOutput)
		wsc.Stop()
	}()
	waitForEvent(t, wsc, eidInput, func() {
		tx := makeDefaultSendTxSigned(t, jsonRpcClient, toAddr, amt)
		broadcastTx(t, jsonRpcClient, tx)
	}, unmarshalValidateSend(amt, toAddr))

	waitForEvent(t, wsc, eidOutput, func() {},
		unmarshalValidateSend(amt, toAddr))
}

// ensure events are only fired once for a given transaction
func TestWSDoubleFire(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	eid := txs.EventStringAccInput(user[0].Address)
	subId := subscribeAndGetSubscriptionId(t, wsc, eid)
	defer func() {
		unsubscribe(t, wsc, subId)
		wsc.Stop()
	}()
	amt := int64(100)
	toAddr := user[1].Address
	// broadcast the transaction, wait to hear about it
	waitForEvent(t, wsc, eid, func() {
		tx := makeDefaultSendTxSigned(t, jsonRpcClient, toAddr, amt)
		broadcastTx(t, jsonRpcClient, tx)
	}, func(eid string, b txs.EventData) (bool, error) {
		return true, nil
	})
	// but make sure we don't hear about it twice
	err := waitForEvent(t, wsc, eid,
		func() {},
		func(eid string, b txs.EventData) (bool, error) {
			return false, nil
		})
	assert.True(t, err.Timeout(), "We should have timed out waiting for second"+
		" %v event", eid)
}

// create a contract, wait for the event, and send it a msg, validate the return
func TestWSCallWait(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	eid1 := txs.EventStringAccInput(user[0].Address)
	subId1 := subscribeAndGetSubscriptionId(t, wsc, eid1)
	defer func() {
		unsubscribe(t, wsc, subId1)
		wsc.Stop()
	}()
	amt, gasLim, fee := int64(10000), int64(1000), int64(1000)
	code, returnCode, returnVal := simpleContract()
	var contractAddr []byte
	// wait for the contract to be created
	waitForEvent(t, wsc, eid1, func() {
		tx := makeDefaultCallTx(t, jsonRpcClient, nil, code, amt, gasLim, fee)
		receipt := broadcastTx(t, jsonRpcClient, tx)
		contractAddr = receipt.ContractAddr
	}, unmarshalValidateTx(amt, returnCode))

	// susbscribe to the new contract
	amt = int64(10001)
	eid2 := txs.EventStringAccOutput(contractAddr)
	subId2 := subscribeAndGetSubscriptionId(t, wsc, eid2)
	defer func() {
		unsubscribe(t, wsc, subId2)
	}()
	// get the return value from a call
	data := []byte{0x1}
	waitForEvent(t, wsc, eid2, func() {
		tx := makeDefaultCallTx(t, jsonRpcClient, contractAddr, data, amt, gasLim, fee)
		receipt := broadcastTx(t, jsonRpcClient, tx)
		contractAddr = receipt.ContractAddr
	}, unmarshalValidateTx(amt, returnVal))
}

// create a contract and send it a msg without waiting. wait for contract event
// and validate return
func TestWSCallNoWait(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	amt, gasLim, fee := int64(10000), int64(1000), int64(1000)
	code, _, returnVal := simpleContract()

	tx := makeDefaultCallTx(t, jsonRpcClient, nil, code, amt, gasLim, fee)
	receipt := broadcastTx(t, jsonRpcClient, tx)
	contractAddr := receipt.ContractAddr

	// susbscribe to the new contract
	amt = int64(10001)
	eid := txs.EventStringAccOutput(contractAddr)
	subId := subscribeAndGetSubscriptionId(t, wsc, eid)
	defer func() {
		unsubscribe(t, wsc, subId)
		wsc.Stop()
	}()
	// get the return value from a call
	data := []byte{0x1}
	waitForEvent(t, wsc, eid, func() {
		tx := makeDefaultCallTx(t, jsonRpcClient, contractAddr, data, amt, gasLim, fee)
		broadcastTx(t, jsonRpcClient, tx)
	}, unmarshalValidateTx(amt, returnVal))
}

// create two contracts, one of which calls the other
func TestWSCallCall(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	amt, gasLim, fee := int64(10000), int64(1000), int64(1000)
	code, _, returnVal := simpleContract()
	txid := new([]byte)

	// deploy the two contracts
	tx := makeDefaultCallTx(t, jsonRpcClient, nil, code, amt, gasLim, fee)
	receipt := broadcastTx(t, jsonRpcClient, tx)
	contractAddr1 := receipt.ContractAddr

	code, _, _ = simpleCallContract(contractAddr1)
	tx = makeDefaultCallTx(t, jsonRpcClient, nil, code, amt, gasLim, fee)
	receipt = broadcastTx(t, jsonRpcClient, tx)
	contractAddr2 := receipt.ContractAddr

	// subscribe to the new contracts
	amt = int64(10001)
	eid := txs.EventStringAccCall(contractAddr1)
	subId := subscribeAndGetSubscriptionId(t, wsc, eid)
	defer func() {
		unsubscribe(t, wsc, subId)
		wsc.Stop()
	}()
	// call contract2, which should call contract1, and wait for ev1

	// let the contract get created first
	waitForEvent(t, wsc, eid, func() {
	}, func(eid string, b txs.EventData) (bool, error) {
		return true, nil
	})
	// call it
	waitForEvent(t, wsc, eid, func() {
		tx := makeDefaultCallTx(t, jsonRpcClient, contractAddr2, nil, amt, gasLim, fee)
		broadcastTx(t, jsonRpcClient, tx)
		*txid = txs.TxHash(chainID, tx)
	}, unmarshalValidateCall(user[0].Address, returnVal, txid))
}

func TestSubscribe(t *testing.T) {
	wsc := newWSClient()
	var subId string
	subscribe(t, wsc, txs.EventStringNewBlock())

	timeout := time.NewTimer(timeoutSeconds * time.Second)
Subscribe:
	for {
		select {
		case <-timeout.C:
			t.Fatal("Timed out waiting for subscription result")

		case bs := <-wsc.ResultsCh:
			resultSubscribe, ok := readResult(t, bs).(*core_types.ResultSubscribe)
			if ok {
				assert.Equal(t, txs.EventStringNewBlock(), resultSubscribe.Event)
				subId = resultSubscribe.SubscriptionId
				break Subscribe
			}
		}
	}

	seenBlock := false
	timeout = time.NewTimer(timeoutSeconds * time.Second)
	for {
		select {
		case <-timeout.C:
			if !seenBlock {
				t.Fatal("Timed out without seeing a NewBlock event")
			}
			return

		case bs := <-wsc.ResultsCh:
			resultEvent, ok := readResult(t, bs).(*core_types.ResultEvent)
			if ok {
				_, ok := resultEvent.Data.(txs.EventDataNewBlock)
				if ok {
					if seenBlock {
						// There's a mild race here, but when we enter we've just seen a block
						// so we should be able to unsubscribe before we see another block
						t.Fatal("Continued to see NewBlock event after unsubscribing")
					} else {
						seenBlock = true
						unsubscribe(t, wsc, subId)
					}
				}
			}
		}
	}
}
