// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package client

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-wire"

	log "github.com/eris-ltd/eris-logger"

	"github.com/eris-ltd/eris-db/txs"
	ctypes "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
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

// NOTE [ben] Compiler check to ensure ErisNodeClient successfully implements
// eris-db/client.NodeClient
var _ NodeWebsocketClient = (*ErisNodeWebsocketClient)(nil)

type ErisNodeWebsocketClient struct {
	// TODO: assert no memory leak on closing with open websocket
	tendermintWebsocket *rpcclient.WSClient
}

// Subscribe to an eventid
func (erisNodeWebsocketClient *ErisNodeWebsocketClient) Subscribe(eventid string) error {
	// TODO we can in the background listen to the subscription id and remember it to ease unsubscribing later.
	return erisNodeWebsocketClient.tendermintWebsocket.Subscribe(eventid)
}

// Unsubscribe from an eventid
func (erisNodeWebsocketClient *ErisNodeWebsocketClient) Unsubscribe(subscriptionId string) error {
	return erisNodeWebsocketClient.tendermintWebsocket.Unsubscribe(subscriptionId)
}

// Returns a channel that will receive a confirmation with a result or the exception that
// has been confirmed; or an error is returned and the confirmation channel is nil.
func (erisNodeWebsocketClient *ErisNodeWebsocketClient) WaitForConfirmation(eventid string) (chan Confirmation, error) {
	// check no errors are reported on the websocket 
	if err := erisNodeWebsocketClient.assertNoErrors(); err != nil {
		return nil, err
	}


	// Setup the confirmation channel to be returned
	confirmationChannel := make(chan Confirmation, 1)
	var latestBlockHash []byte
	var inputAddr []byte
	var chainId string
	var tx txs.Tx
	eid := txs.EventStringAccInput(inputAddr)

	// Read the incoming events
	go func() {
		var err error
		for {
			resultBytes := <- erisNodeWebsocketClient.tendermintWebsocket.ResultsCh
			result := new(ctypes.ErisDBResult)
			if wire.ReadJSONPtr(result, resultBytes, &err); err != nil {
				// keep calm and carry on
				log.Errorf("eris-client - Failed to unmarshal json bytes for websocket event: %s", err)
				continue
			}
			
			event, ok := (*result).(*ctypes.ResultEvent)
			if !ok {
				// keep calm and carry on
				log.Error("eris-client - Failed to cast to ResultEvent for websocket event")
				continue
			}
			
			blockData, ok := event.Data.(txs.EventDataNewBlock)
			if ok {
				latestBlockHash = blockData.Block.Hash()
				log.WithFields(log.Fields{
					"new block": blockData.Block,
					"latest hash": latestBlockHash,
				}).Debug("Registered new block")
				continue
			}
			
			// we don't accept events unless they came after a new block (ie. in)
			if latestBlockHash == nil {
				continue
			}

			if event.Event != eid {
				log.Warnf("Received unsolicited event! Got %s, expected %s\n", event.Event, eid)
				continue
			}

			data, ok := event.Data.(txs.EventDataTx)
			if !ok {
				// We are on the lookout for EventDataTx
				confirmationChannel <- Confirmation{
					BlockHash: latestBlockHash,
					Event: nil,
					Exception: fmt.Errorf("response error: expected result.Data to be *types.EventDataTx"),
					Error: nil,
				}
				return
			}

			if !bytes.Equal(txs.TxHash(chainId, data.Tx), txs.TxHash(chainId, tx)) {
				log.WithFields(log.Fields{
					// TODO: consider re-implementing TxID again, or other more clear debug
					"received transaction event": txs.TxHash(chainId, data.Tx),
				}).Debug("Received different event")
				continue
			}

			if data.Exception != "" {
				confirmationChannel <- Confirmation{
					BlockHash: latestBlockHash,
					Event: &data,
					Exception: fmt.Errorf("Transaction confirmed with exception:", data.Exception),
					Error: nil,
				}
				return
			}
			
			// success, return the full event and blockhash and exit go-routine
			confirmationChannel <- Confirmation{
				BlockHash: latestBlockHash,
				Event: &data,
				Exception: nil,
				Error: nil,
			}
			return
		}

	}()

	// TODO: [ben] this is a draft implementation as resources on time.After can not be
	// recovered before the timeout.  Close-down timeout at success properly.
	timeout := time.After(time.Duration(MaxCommitWaitTimeSeconds) * time.Second)

	go func() {
		<- timeout
		confirmationChannel <- Confirmation{
			BlockHash: nil,
			Event: nil,
			Exception: nil,
			Error: fmt.Errorf("timed out waiting for event"),
		}
		return
	}()
	return confirmationChannel, nil
}

func (erisNodeWebsocketClient *ErisNodeWebsocketClient) Close() {
	if erisNodeWebsocketClient.tendermintWebsocket != nil {
		erisNodeWebsocketClient.tendermintWebsocket.Stop()
	}
}

func (erisNodeWebsocketClient *ErisNodeWebsocketClient) assertNoErrors() error {
	if erisNodeWebsocketClient.tendermintWebsocket != nil {
		select {
		case err := <-erisNodeWebsocketClient.tendermintWebsocket.ErrorsCh:
			return err
		default:
			return nil
		}
	} else {
		return fmt.Errorf("Eris-client has no websocket initialised.")
	}
}