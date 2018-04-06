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
	"context"

	"github.com/hyperledger/burrow/rpc/tm"
	"github.com/tendermint/tendermint/rpc/lib/types"
)

type WebsocketClient interface {
	Send(ctx context.Context, request rpctypes.RPCRequest) error
}

const SubscribeRequestID = "Subscribe"
const UnsubscribeRequestID = "Unsubscribe"

func EventResponseID(eventID string) string {
	return tm.EventResponseID(SubscribeRequestID, eventID)
}

func Subscribe(wsc WebsocketClient, eventID string) error {
	req, err := rpctypes.MapToRequest(SubscribeRequestID,
		"subscribe", map[string]interface{}{"eventID": eventID})
	if err != nil {
		return err
	}
	return wsc.Send(context.Background(), req)
}

func Unsubscribe(websocketClient WebsocketClient, subscriptionID string) error {
	req, err := rpctypes.MapToRequest(UnsubscribeRequestID,
		"unsubscribe", map[string]interface{}{"subscriptionID": subscriptionID})
	if err != nil {
		return err
	}
	return websocketClient.Send(context.Background(), req)
}
