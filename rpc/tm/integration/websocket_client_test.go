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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/rpc"
	tmClient "github.com/hyperledger/burrow/rpc/tm/client"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmTypes "github.com/tendermint/tendermint/types"
)

//--------------------------------------------------------------------------------
// Test the websocket service

// make a simple connection to the server
func TestWSConnect(t *testing.T) {
	wsc := newWSClient()
	stopWSClient(wsc)
}

// receive a new block message
func TestWSNewBlock(t *testing.T) {
	wsc := newWSClient()
	eid := tmTypes.EventNewBlock
	subId := subscribeAndGetSubscriptionId(t, wsc, eid)
	defer func() {
		unsubscribe(t, wsc, subId)
		stopWSClient(wsc)
	}()
	waitForEvent(t, wsc, eid, func() {},
		func(eventID string, resultEvent *rpc.ResultEvent) (bool, error) {
			fmt.Println("Check: ", resultEvent.Tendermint.EventDataNewBlock().Block)
			return true, nil
		})
}

// receive a few new block messages in a row, with increasing height
func TestWSBlockchainGrowth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	eid := tmTypes.EventNewBlock
	subId := subscribeAndGetSubscriptionId(t, wsc, eid)
	defer func() {
		unsubscribe(t, wsc, subId)
		stopWSClient(wsc)
	}()
	// listen for NewBlock, ensure height increases by 1
	var initBlockN int64
	for i := int64(0); i < 2; i++ {
		waitForEvent(t, wsc, eid, func() {},
			func(eventID string, resultEvent *rpc.ResultEvent) (bool, error) {
				eventDataNewBlock := resultEvent.Tendermint.EventDataNewBlock()
				if eventDataNewBlock == nil {
					t.Fatalf("Was expecting EventDataNewBlock but got %v", resultEvent)
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
	toAddr := privateAccounts[1].Address()
	amt := uint64(100)
	eidInput := events.EventStringAccountInput(privateAccounts[0].Address())
	eidOutput := events.EventStringAccountOutput(toAddr)
	subIdInput := subscribeAndGetSubscriptionId(t, wsc, eidInput)
	subIdOutput := subscribeAndGetSubscriptionId(t, wsc, eidOutput)
	defer func() {
		unsubscribe(t, wsc, subIdInput)
		unsubscribe(t, wsc, subIdOutput)
		stopWSClient(wsc)
	}()

	tx := makeDefaultSendTxSigned(t, jsonRpcClient, toAddr, amt)
	broadcastTx(t, jsonRpcClient, tx)

	// Set of response IDs we expect
	rids := map[string]struct{}{tmClient.EventResponseID(eidInput): {}, tmClient.EventResponseID(eidOutput): {}}

	r := <-wsc.ResponsesCh
	result, err := readResponse(r)
	assert.NoError(t, err)
	assert.NoError(t, unmarshalValidateSend(amt, toAddr, result))
	delete(rids, r.ID)

	r = <-wsc.ResponsesCh
	result, err = readResponse(r)
	assert.NoError(t, err)
	assert.NoError(t, unmarshalValidateSend(amt, toAddr, result))
	delete(rids, r.ID)

	// Note currently we cannot guarantee order with pubsub event system
	assert.Empty(t, rids, "Should receive input and output event")
}

// ensure events are only fired once for a given transaction
func TestWSDoubleFire(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	eid := events.EventStringAccountInput(privateAccounts[0].Address())
	subId := subscribeAndGetSubscriptionId(t, wsc, eid)
	defer func() {
		unsubscribe(t, wsc, subId)
		stopWSClient(wsc)
	}()
	amt := uint64(100)
	toAddr := privateAccounts[1].Address()
	// broadcast the transaction, wait to hear about it
	waitForEvent(t, wsc, eid, func() {
		tx := makeDefaultSendTxSigned(t, jsonRpcClient, toAddr, amt)
		broadcastTx(t, jsonRpcClient, tx)
	}, func(eventID string, resultEvent *rpc.ResultEvent) (bool, error) {
		return true, nil
	})
	select {
	case <-wsc.ResponsesCh:
		t.Fatal("Should time out waiting for second event")
	case <-time.After(timeoutSeconds * time.Second):

	}
}

// create a contract, wait for the event, and send it a msg, validate the return
func TestWSCallWait(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	defer stopWSClient(wsc)
	// Mini soak test
	for i := 0; i < 20; i++ {
		amt, gasLim, fee := uint64(10000), uint64(1000), uint64(1000)
		code, returnCode, returnVal := simpleContract()
		var contractAddr crypto.Address
		eid1 := events.EventStringAccountInput(privateAccounts[0].Address())
		subId1 := subscribeAndGetSubscriptionId(t, wsc, eid1)
		// wait for the contract to be created
		waitForEvent(t, wsc, eid1, func() {
			tx := makeDefaultCallTx(t, jsonRpcClient, nil, code, amt, gasLim, fee)
			receipt := broadcastTx(t, jsonRpcClient, tx)
			contractAddr = receipt.ContractAddress
		}, unmarshalValidateTx(amt, returnCode))

		unsubscribe(t, wsc, subId1)

		// susbscribe to the new contract
		amt = uint64(10001)
		eid2 := events.EventStringAccountOutput(contractAddr)
		subId2 := subscribeAndGetSubscriptionId(t, wsc, eid2)
		// get the return value from a call
		data := []byte{0x1}
		waitForEvent(t, wsc, eid2, func() {
			tx := makeDefaultCallTx(t, jsonRpcClient, &contractAddr, data, amt, gasLim, fee)
			receipt := broadcastTx(t, jsonRpcClient, tx)
			contractAddr = receipt.ContractAddress
		}, unmarshalValidateTx(amt, returnVal))
		unsubscribe(t, wsc, subId2)
	}
}

// create a contract and send it a msg without waiting. wait for contract event
// and validate return
func TestWSCallNoWait(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	defer stopWSClient(wsc)
	amt, gasLim, fee := uint64(10000), uint64(1000), uint64(1000)
	code, _, returnVal := simpleContract()

	sequence := getSequence(t, jsonRpcClient, privateAccounts[0].Address())
	txEnv := makeDefaultCallTx(t, jsonRpcClient, nil, code, amt, gasLim, fee)
	receipt := broadcastTx(t, jsonRpcClient, txEnv)
	contractAddr := receipt.ContractAddress

	// susbscribe to the new contract
	amt = uint64(10001)
	eid := events.EventStringAccountOutput(contractAddr)
	subId := subscribeAndGetSubscriptionId(t, wsc, eid)
	defer unsubscribe(t, wsc, subId)

	data := []byte{0x1}
	waitForEvent(t, wsc, eid, func() {
		tx := payload.NewCallTxWithSequence(privateAccounts[0].PublicKey(), &contractAddr, data, amt, gasLim, fee,
			sequence+3)
		txEnv = txs.Enclose(genesisDoc.ChainID(), tx)
		require.NoError(t, txEnv.Sign(privateAccounts[0]))
		broadcastTx(t, jsonRpcClient, txEnv)
	}, unmarshalValidateTx(amt, returnVal))
}

// create two contracts, one of which calls the other
func TestWSCallCall(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	defer stopWSClient(wsc)
	amt, gasLim, fee := uint64(10000), uint64(1000), uint64(1000)
	code, _, returnVal := simpleContract()
	TxHash := new([]byte)

	// deploy the two contracts
	txEnv := makeDefaultCallTx(t, jsonRpcClient, nil, code, amt, gasLim, fee)
	receipt := txEnv.Tx.GenerateReceipt()
	contractAddr1 := receipt.ContractAddress
	// subscribe to the new contracts
	eid := events.EventStringAccountCall(contractAddr1)
	subId := subscribeAndGetSubscriptionId(t, wsc, eid)
	defer unsubscribe(t, wsc, subId)

	_, err := broadcastTxAndWait(t, jsonRpcClient, txEnv)
	require.NoError(t, err)

	// call contract2, which should call contract1, and wait for ev1
	code, _, _ = simpleCallContract(contractAddr1)
	txEnv = makeDefaultCallTx(t, jsonRpcClient, nil, code, amt, gasLim, fee)
	receipt2, err := broadcastTxAndWait(t, jsonRpcClient, txEnv)
	require.NoError(t, err)
	contractAddr2 := receipt2.ContractAddress

	// let the contract get created first
	waitForEvent(t, wsc, eid,
		// Runner
		func() {
		},
		// Event Checker
		func(eventID string, resultEvent *rpc.ResultEvent) (bool, error) {
			return true, nil
		})
	// call it
	waitForEvent(t, wsc, eid,
		// Runner
		func() {
			txEnv := makeDefaultCallTx(t, jsonRpcClient, &contractAddr2, nil, amt, gasLim, fee)
			broadcastTx(t, jsonRpcClient, txEnv)
			*TxHash = txEnv.Tx.Hash()
		},
		// Event checker
		unmarshalValidateCall(privateAccounts[0].Address(), returnVal, TxHash))
}

func TestSubscribe(t *testing.T) {
	wsc := newWSClient()
	var subId string
	subscribe(t, wsc, tmTypes.EventNewBlock)

Subscribe:
	for {
		select {
		case <-time.After(timeoutSeconds * time.Second):
			t.Fatal("Timed out waiting for subscription result")

		case response := <-wsc.ResponsesCh:
			require.Nil(t, response.Error)
			res := new(rpc.ResultSubscribe)
			require.NoError(t, json.Unmarshal(response.Result, res))
			assert.Equal(t, tmTypes.EventNewBlock, res.EventID)
			subId = res.SubscriptionID
			break Subscribe
		}
	}

	blocksSeen := 0
	for {
		select {
		// wait long enough to check we don't see another new block event even though
		// a block will have come
		case <-time.After(expectBlockInSeconds * time.Second):
			if blocksSeen == 0 {
				t.Fatal("Timed out without seeing a NewBlock event")
			}
			return

		case response := <-wsc.ResponsesCh:
			require.Nil(t, response.Error)

			if response.ID == tmClient.EventResponseID(tmTypes.EventNewBlock) {
				res := new(rpc.ResultEvent)
				json.Unmarshal(response.Result, res)
				enb := res.Tendermint.EventDataNewBlock()
				if enb != nil {
					assert.Equal(t, genesisDoc.ChainID(), enb.Block.ChainID)
					if blocksSeen > 1 {
						t.Fatal("Continued to see NewBlock event after unsubscribing")
					} else {
						if blocksSeen == 0 {
							unsubscribe(t, wsc, subId)
						}
						blocksSeen++
					}
				}
			}
		}
	}
}
