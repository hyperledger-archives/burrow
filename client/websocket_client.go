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

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	exeEvents "github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/tm/client"
	rpcClient "github.com/hyperledger/burrow/rpc/tm/lib/client"
	"github.com/hyperledger/burrow/txs"
	tmTypes "github.com/tendermint/tendermint/types"
)

const (
	MaxCommitWaitTimeSeconds = 10
)

type Confirmation struct {
	BlockHash   []byte
	EventDataTx *exeEvents.EventDataTx
	Exception   error
	Error       error
}

// NOTE [ben] Compiler check to ensure burrowNodeClient successfully implements
// burrow/client.NodeClient
var _ NodeWebsocketClient = (*burrowNodeWebsocketClient)(nil)

type burrowNodeWebsocketClient struct {
	// TODO: assert no memory leak on closing with open websocket
	tendermintWebsocket *rpcClient.WSClient
	logger              *logging.Logger
}

// Subscribe to an eventid
func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) Subscribe(eventId string) error {
	// TODO we can in the background listen to the subscription id and remember it to ease unsubscribing later.
	return client.Subscribe(burrowNodeWebsocketClient.tendermintWebsocket,
		eventId)
}

// Unsubscribe from an eventid
func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) Unsubscribe(subscriptionId string) error {
	return client.Unsubscribe(burrowNodeWebsocketClient.tendermintWebsocket,
		subscriptionId)
}

// Returns a channel that will receive a confirmation with a result or the exception that
// has been confirmed; or an error is returned and the confirmation channel is nil.
func (burrowNodeWebsocketClient *burrowNodeWebsocketClient) WaitForConfirmation(txEnv *txs.Envelope,
	inputAddr crypto.Address) (chan Confirmation, error) {

	// Setup the confirmation channel to be returned
	confirmationChannel := make(chan Confirmation, 1)
	var latestBlockHash []byte

	eventID := exeEvents.EventStringAccountInput(inputAddr)
	if err := burrowNodeWebsocketClient.Subscribe(eventID); err != nil {
		return nil, fmt.Errorf("Error subscribing to AccInput event (%s): %v", eventID, err)
	}
	if err := burrowNodeWebsocketClient.Subscribe(tmTypes.EventNewBlock); err != nil {
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
					burrowNodeWebsocketClient.logger.InfoMsg(
						"Error received on websocket channel", structure.ErrorKey, err)
					continue
				}

				switch response.ID {
				case client.SubscribeRequestID:
					resultSubscribe := new(rpc.ResultSubscribe)
					err = json.Unmarshal(response.Result, resultSubscribe)
					if err != nil {
						burrowNodeWebsocketClient.logger.InfoMsg("Unable to unmarshal ResultSubscribe",
							structure.ErrorKey, err)
						continue
					}
					// TODO: collect subscription IDs, push into channel and on completion
					burrowNodeWebsocketClient.logger.InfoMsg("Received confirmation for event",
						"event", resultSubscribe.EventID,
						"subscription_id", resultSubscribe.SubscriptionID)

				case client.EventResponseID(tmTypes.EventNewBlock):
					resultEvent := new(rpc.ResultEvent)
					err = json.Unmarshal(response.Result, resultEvent)
					if err != nil {
						burrowNodeWebsocketClient.logger.InfoMsg("Unable to unmarshal ResultEvent",
							structure.ErrorKey, err)
						continue
					}
					blockData := resultEvent.Tendermint.EventDataNewBlock()
					if blockData != nil {
						latestBlockHash = blockData.Block.Hash()
						burrowNodeWebsocketClient.logger.TraceMsg("Registered new block",
							"block", blockData.Block,
							"latest_block_hash", latestBlockHash,
						)
					}

				case client.EventResponseID(eventID):
					resultEvent := new(rpc.ResultEvent)
					err = json.Unmarshal(response.Result, resultEvent)
					if err != nil {
						burrowNodeWebsocketClient.logger.InfoMsg("Unable to unmarshal ResultEvent",
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

					if !bytes.Equal(eventDataTx.Tx.Hash(), txEnv.Tx.Hash()) {
						burrowNodeWebsocketClient.logger.TraceMsg("Received different event",
							// TODO: consider re-implementing TxID again, or other more clear debug
							"received transaction event", eventDataTx.Tx.Hash())
						continue
					}

					if eventDataTx.Exception != nil {
						confirmationChannel <- Confirmation{
							BlockHash:   latestBlockHash,
							EventDataTx: eventDataTx,
							Exception: errors.Wrap(eventDataTx.Exception,
								"transaction confirmed but execution gave exception: %v"),
							Error: nil,
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
					burrowNodeWebsocketClient.logger.InfoMsg("Received unsolicited response",
						"response_id", response.ID,
						"expected_response_id", client.EventResponseID(eventID))
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
