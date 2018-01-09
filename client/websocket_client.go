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

	"encoding/json"

	"github.com/hyperledger/burrow/account"
	exe_events "github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/rpc"
	tendermint_client "github.com/hyperledger/burrow/rpc/tm/client"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/rpc/lib/client"
	tm_types "github.com/tendermint/tendermint/types"
)

const (
	MaxCommitWaitTimeSeconds = 10
)

type Confirmation struct {
	BlockHash   []byte
	EventDataTx *exe_events.EventDataTx
	Exception   error
	Error       error
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
func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) WaitForConfirmation(tx txs.Tx, chainId string,
	inputAddr account.Address) (chan Confirmation, error) {

	// Setup the confirmation channel to be returned
	confirmationChannel := make(chan Confirmation, 1)
	var latestBlockHash []byte

	eid := exe_events.EventStringAccInput(inputAddr)
	if err := burrowNodeWebsocketClient.Subscribe(eid); err != nil {
		return nil, fmt.Errorf("Error subscribing to AccInput event (%s): %v", eid, err)
	}
	if err := burrowNodeWebsocketClient.Subscribe(tm_types.EventStringNewBlock()); err != nil {
		return nil, fmt.Errorf("Error subscribing to NewBlock event: %v", err)
	}
	// Read the incoming events
	go func() {
		var err error
		for {
			response := <-burrowNodeWebsocketClient.tendermintWebsocket.ResponsesCh
			if response.Error != nil {
				logging.InfoMsg(burrowNodeWebsocketClient.logger,
					"Error received on websocket channel", "error", err)
				continue
			}
			result := new(rpc.Result)

			if json.Unmarshal(*response.Result, result); err != nil {
				// keep calm and carry on
				logging.InfoMsg(burrowNodeWebsocketClient.logger,
					"Failed to unmarshal json bytes for websocket event", "error", err)
				continue
			}

			subscription, ok := result.Unwrap().(*rpc.ResultSubscribe)
			if ok {
				// Received confirmation of subscription to event streams
				// TODO: collect subscription IDs, push into channel and on completion
				// unsubscribe
				logging.InfoMsg(burrowNodeWebsocketClient.logger, "Received confirmation for event",
					"event", subscription.Event,
					"subscription_id", subscription.SubscriptionId)
				continue
			}

			resultEvent, ok := result.Unwrap().(*rpc.ResultEvent)
			if !ok {
				// keep calm and carry on
				logging.InfoMsg(burrowNodeWebsocketClient.logger, "Failed to cast to ResultEvent for websocket event",
					"event", resultEvent.Event)
				continue
			}

			blockData := resultEvent.EventDataNewBlock()
			if blockData != nil {
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

			if resultEvent.Event != eid {
				logging.InfoMsg(burrowNodeWebsocketClient.logger, "Received unsolicited event",
					"event_received", resultEvent.Event,
					"event_expected", eid)
				continue
			}

			eventDataTx := resultEvent.EventDataTx()
			if eventDataTx == nil {
				// We are on the lookout for EventDataTx
				confirmationChannel <- Confirmation{
					BlockHash:   latestBlockHash,
					EventDataTx: nil,
					Exception:   fmt.Errorf("response error: expected result.Data to be *types.EventDataTx"),
					Error:       nil,
				}
				return
			}

			if !bytes.Equal(txs.TxHash(chainId, eventDataTx.Tx), txs.TxHash(chainId, tx)) {
				logging.TraceMsg(burrowNodeWebsocketClient.logger, "Received different event",
					// TODO: consider re-implementing TxID again, or other more clear debug
					"received transaction event", txs.TxHash(chainId, eventDataTx.Tx))
				continue
			}

			if eventDataTx.Exception != "" {
				confirmationChannel <- Confirmation{
					BlockHash:   latestBlockHash,
					EventDataTx: eventDataTx,
					Exception:   fmt.Errorf("Transaction confirmed with exception: %v", eventDataTx.Exception),
					Error:       nil,
				}
				return
			}
			// success, return the full event and blockhash and exit go-routine
			confirmationChannel <- Confirmation{
				BlockHash:   latestBlockHash,
				EventDataTx: eventDataTx,
				Exception:   nil,
				Error:       nil,
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
			BlockHash:   nil,
			EventDataTx: nil,
			Exception:   nil,
			Error:       fmt.Errorf("timed out waiting for event"),
		}
	}()
	return confirmationChannel, nil
}

func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) Close() {
	if burrowNodeWebsocketClient.tendermintWebsocket != nil {
		burrowNodeWebsocketClient.tendermintWebsocket.Stop()
	}
}
