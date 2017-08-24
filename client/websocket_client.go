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
	"fmt"
	"time"

	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-wire"

	"github.com/hyperledger/burrow/logging"
	tendermint_client "github.com/hyperledger/burrow/rpc/tendermint/client"
	ctypes "github.com/hyperledger/burrow/rpc/tendermint/core/types"
	"github.com/hyperledger/burrow/txs"
)

const (
	MaxCommitWaitTimeSeconds = 10
)

type Confirmation struct {
	BlockHash []byte
	Event     txs.EventData
	Exception error
	Error     error
}

// NOTE [ben] Compiler check to ensure burrowNodeClient successfully implements
// burrow/client.NodeClient
var _ NodeWebsocketClient = (*burrowNodeWebsocketClient)(nil)

type burrowNodeWebsocketClient struct {
	// TODO: assert no memory leak on closing with open websocket
	tendermintWebsocket *rpcclient.WSClient
	logger              logging_types.InfoTraceLogger
}

// Subscribe to an eventid
func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) Subscribe(eventId string) error {
	// TODO we can in the background listen to the subscription id and remember it to ease unsubscribing later.
	return tendermint_client.Subscribe(burrowNodeWebsocketClient.tendermintWebsocket,
		eventId)
}

// Unsubscribe from an eventid
func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) Unsubscribe(subscriptionId string) error {
	return tendermint_client.Unsubscribe(burrowNodeWebsocketClient.tendermintWebsocket,
		subscriptionId)
}

// Returns a channel that will receive a confirmation with a result or the exception that
// has been confirmed; or an error is returned and the confirmation channel is nil.
func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) WaitForConfirmation(tx txs.Tx, chainId string, inputAddr []byte) (chan Confirmation, error) {
	// check no errors are reported on the websocket
	if err := burrowNodeWebsocketClient.assertNoErrors(); err != nil {
		return nil, err
	}

	// Setup the confirmation channel to be returned
	confirmationChannel := make(chan Confirmation, 1)
	var latestBlockHash []byte

	eid := txs.EventStringAccInput(inputAddr)
	if err := burrowNodeWebsocketClient.Subscribe(eid); err != nil {
		return nil, fmt.Errorf("Error subscribing to AccInput event (%s): %v", eid, err)
	}
	if err := burrowNodeWebsocketClient.Subscribe(txs.EventStringNewBlock()); err != nil {
		return nil, fmt.Errorf("Error subscribing to NewBlock event: %v", err)
	}
	// Read the incoming events
	go func() {
		var err error
		for {
			resultBytes := <-burrowNodeWebsocketClient.tendermintWebsocket.ResultsCh
			result := new(ctypes.BurrowResult)
			if wire.ReadJSONPtr(result, resultBytes, &err); err != nil {
				// keep calm and carry on
				logging.InfoMsg(burrowNodeWebsocketClient.logger, "Failed to unmarshal json bytes for websocket event",
					"error", err)
				continue
			}

			subscription, ok := (*result).(*ctypes.ResultSubscribe)
			if ok {
				// Received confirmation of subscription to event streams
				// TODO: collect subscription IDs, push into channel and on completion
				// unsubscribe
				logging.InfoMsg(burrowNodeWebsocketClient.logger, "Received confirmation for event",
					"event", subscription.Event,
					"subscription_id", subscription.SubscriptionId)
				continue
			}

			event, ok := (*result).(*ctypes.ResultEvent)
			if !ok {
				// keep calm and carry on
				logging.InfoMsg(burrowNodeWebsocketClient.logger, "Failed to cast to ResultEvent for websocket event",
					"event", event.Event)
				continue
			}

			blockData, ok := event.Data.(txs.EventDataNewBlock)
			if ok {
				latestBlockHash = blockData.Block.Hash()
				logging.TraceMsg(burrowNodeWebsocketClient.logger, "Registered new block",
					"block", blockData.Block,
					"latest_block_hash", latestBlockHash,
				)
				continue
			}

			// NOTE: [ben] hotfix on 0.16.1 because NewBlock events to arrive seemingly late
			// we now miss events because of this redundant check;  This check is safely removed
			// because for CallTx on checking the transaction is not run in the EVM and no false
			// positive event can be sent; neither is this check a good check for that.

			// we don't accept events unless they came after a new block (ie. in)
			// if latestBlockHash == nil {
			// 	logging.InfoMsg(burrowNodeWebsocketClient.logger, "First block has not been registered so ignoring event",
			// 		"event", event.Event)
			// 	continue
			// }

			if event.Event != eid {
				logging.InfoMsg(burrowNodeWebsocketClient.logger, "Received unsolicited event",
					"event_received", event.Event,
					"event_expected", eid)
				continue
			}

			data, ok := event.Data.(txs.EventDataTx)
			if !ok {
				// We are on the lookout for EventDataTx
				confirmationChannel <- Confirmation{
					BlockHash: latestBlockHash,
					Event:     nil,
					Exception: fmt.Errorf("response error: expected result.Data to be *types.EventDataTx"),
					Error:     nil,
				}
				return
			}

			if !bytes.Equal(txs.TxHash(chainId, data.Tx), txs.TxHash(chainId, tx)) {
				logging.TraceMsg(burrowNodeWebsocketClient.logger, "Received different event",
					// TODO: consider re-implementing TxID again, or other more clear debug
					"received transaction event", txs.TxHash(chainId, data.Tx))
				continue
			}

			if data.Exception != "" {
				confirmationChannel <- Confirmation{
					BlockHash: latestBlockHash,
					Event:     &data,
					Exception: fmt.Errorf("Transaction confirmed with exception: %v", data.Exception),
					Error:     nil,
				}
				return
			}
			// success, return the full event and blockhash and exit go-routine
			confirmationChannel <- Confirmation{
				BlockHash: latestBlockHash,
				Event:     &data,
				Exception: nil,
				Error:     nil,
			}
			return
		}

	}()

	// TODO: [ben] this is a draft implementation as resources on time.After can not be
	// recovered before the timeout.  Close-down timeout at success properly.
	timeout := time.After(time.Duration(MaxCommitWaitTimeSeconds) * time.Second)

	go func() {
		<-timeout
		confirmationChannel <- Confirmation{
			BlockHash: nil,
			Event:     nil,
			Exception: nil,
			Error:     fmt.Errorf("timed out waiting for event"),
		}
		return
	}()
	return confirmationChannel, nil
}

func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) Close() {
	if burrowNodeWebsocketClient.tendermintWebsocket != nil {
		burrowNodeWebsocketClient.tendermintWebsocket.Stop()
	}
}

func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) assertNoErrors() error {
	if burrowNodeWebsocketClient.tendermintWebsocket != nil {
		select {
		case err := <-burrowNodeWebsocketClient.tendermintWebsocket.ErrorsCh:
			return err
		default:
			return nil
		}
	} else {
		return fmt.Errorf("burrow-client has no websocket initialised.")
	}
}
