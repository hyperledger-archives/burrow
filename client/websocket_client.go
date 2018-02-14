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
	"github.com/hyperledger/burrow/logging/structure"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/rpc"
	tm_client "github.com/hyperledger/burrow/rpc/tm/client"
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
	return tm_client.Subscribe(burrowNodeWebsocketClient.tendermintWebsocket,
		eventId)
}

// Unsubscribe from an eventid
func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) Unsubscribe(subscriptionId string) error {
	return tm_client.Unsubscribe(burrowNodeWebsocketClient.tendermintWebsocket,
		subscriptionId)
}

// Returns a channel that will receive a confirmation with a result or the exception that
// has been confirmed; or an error is returned and the confirmation channel is nil.
func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) WaitForConfirmation(tx txs.Tx, chainId string,
	inputAddr account.Address) (chan Confirmation, error) {

	// Setup the confirmation channel to be returned
	confirmationChannel := make(chan Confirmation, 1)
	var latestBlockHash []byte

	eventID := exe_events.EventStringAccountInput(inputAddr)
	if err := burrowNodeWebsocketClient.Subscribe(eventID); err != nil {
		return nil, fmt.Errorf("Error subscribing to AccInput event (%s): %v", eventID, err)
	}
	if err := burrowNodeWebsocketClient.Subscribe(tm_types.EventNewBlock); err != nil {
		return nil, fmt.Errorf("Error subscribing to NewBlock event: %v", err)
	}
	// Read the incoming events
	go func() {
		var err error

		timeoutTimer := time.NewTimer(time.Duration(MaxCommitWaitTimeSeconds) * time.Second)
		defer func() {
			if !timeoutTimer.Stop() {
				<-timeoutTimer.C
			}
		}()

		for {
			select {
			case <-timeoutTimer.C:
				confirmationChannel <- Confirmation{
					BlockHash:   nil,
					EventDataTx: nil,
					Exception:   nil,
					Error:       fmt.Errorf("timed out waiting for event"),
				}
				return

			case response := <-burrowNodeWebsocketClient.tendermintWebsocket.ResponsesCh:
				if response.Error != nil {
					logging.InfoMsg(burrowNodeWebsocketClient.logger,
						"Error received on websocket channel", structure.ErrorKey, err)
					continue
				}

				switch response.ID {
				case tm_client.SubscribeRequestID:
					resultSubscribe := new(rpc.ResultSubscribe)
					err = json.Unmarshal(response.Result, resultSubscribe)
					if err != nil {
						logging.InfoMsg(burrowNodeWebsocketClient.logger, "Unable to unmarshal ResultSubscribe",
							structure.ErrorKey, err)
						continue
					}
					// TODO: collect subscription IDs, push into channel and on completion
					logging.InfoMsg(burrowNodeWebsocketClient.logger, "Received confirmation for event",
						"event", resultSubscribe.EventID,
						"subscription_id", resultSubscribe.SubscriptionID)

				case tm_client.EventResponseID(tm_types.EventNewBlock):
					resultEvent := new(rpc.ResultEvent)
					err = json.Unmarshal(response.Result, resultEvent)
					if err != nil {
						logging.InfoMsg(burrowNodeWebsocketClient.logger, "Unable to unmarshal ResultEvent",
							structure.ErrorKey, err)
						continue
					}
					blockData := resultEvent.EventDataNewBlock()
					if blockData != nil {
						latestBlockHash = blockData.Block.Hash()
						logging.TraceMsg(burrowNodeWebsocketClient.logger, "Registered new block",
							"block", blockData.Block,
							"latest_block_hash", latestBlockHash,
						)
					}

				case tm_client.EventResponseID(eventID):
					resultEvent := new(rpc.ResultEvent)
					err = json.Unmarshal(response.Result, resultEvent)
					if err != nil {
						logging.InfoMsg(burrowNodeWebsocketClient.logger, "Unable to unmarshal ResultEvent",
							structure.ErrorKey, err)
						continue
					}

					eventDataTx := resultEvent.EventDataTx
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
							Exception:   fmt.Errorf("transaction confirmed with exception: %v", eventDataTx.Exception),
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

				default:
					logging.InfoMsg(burrowNodeWebsocketClient.logger, "Received unsolicited response",
						"response_id", response.ID,
						"expected_response_id", tm_client.EventResponseID(eventID))
				}
			}
		}

	}()

	return confirmationChannel, nil
}

func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) Close() {
	if burrowNodeWebsocketClient.tendermintWebsocket != nil {
		burrowNodeWebsocketClient.tendermintWebsocket.Stop()
	}
}
