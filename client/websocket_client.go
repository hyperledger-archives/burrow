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
	"encoding/json"
	"fmt"

	"github.com/tendermint/go-rpc/client"

	"github.com/eris-ltd/eris-bd/txs"
)

type Confirmation {
	BlockHash []byte
	Event     *txs.EventData
	Exception error
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
	return erisNodeWebsocketClient.tendermintWebsocket.Subscribe(eventid)
}

// Unsubscribe from an eventid
func (erisNodeWebsocketClient *ErisNodeWebsocketClient) Unsubscribe(eventid string) error {
	return erisNodeWebsocketClient.tendermintWebsocket.Unsubscribe(eventid)
}

// Returns a channel with a sign
func (erisNodeWebsocketClient *ErisNodeWebsocketClient) Wait(eventid string) (chan Confirmation, error) {
	// Setup the confirmation channel to be returned
	confirmationChannel := make(chan Confirmation, 1)
	var latestBlockHash []byte

}

func (erisNodeWebsocketClient *ErisNodeWebsocketClient) Close() {
	if erisNodeWebsocketClient.tendermintWebsocket != nil {
		erisNodeWebsocketClient.tendermintWebsocket.Stop()
	}
}

func (erisNodeWebsocketClient *ErisNodeWebsocketClient) assertNoErrors() error {
	if erisNodeWebsocketClient.tendermintWebsocket != nil {

	} else {
		return fmt.Errorf("")
	}
}