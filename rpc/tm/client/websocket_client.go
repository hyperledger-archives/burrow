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

import "context"

type WebsocketClient interface {
	Call(ctx context.Context, method string, params map[string]interface{}) error
}

func Subscribe(websocketClient WebsocketClient, eventId string) error {
	return websocketClient.Call(context.Background(), "subscribe",
		map[string]interface{}{"eventId": eventId})
}

func Unsubscribe(websocketClient WebsocketClient, subscriptionId string) error {
	return websocketClient.Call(context.Background(), "unsubscribe",
		map[string]interface{}{"subscriptionId": subscriptionId})
}
