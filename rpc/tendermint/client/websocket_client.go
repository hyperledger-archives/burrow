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

import "github.com/tendermint/go-rpc/types"

type WebsocketClient interface {
	WriteJSON(v interface{}) error
}

func Subscribe(websocketClient WebsocketClient, eventId string) error {
	return websocketClient.WriteJSON(rpctypes.RPCRequest{
		JSONRPC: "2.0",
		ID:      "",
		Method:  "subscribe",
		Params:  map[string]interface{}{"eventId": eventId},
	})
}

func Unsubscribe(websocketClient WebsocketClient, subscriptionId string) error {
	return websocketClient.WriteJSON(rpctypes.RPCRequest{
		JSONRPC: "2.0",
		ID:      "",
		Method:  "unsubscribe",
		Params:  map[string]interface{}{"subscriptionId": subscriptionId},
	})
}
