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

package events

import (
	"context"
	"fmt"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/tmthrgd/go-hex"
)

// Functions to generate eventId strings

func EventStringAccountCall(addr crypto.Address) string { return fmt.Sprintf("Acc/%s/Call", addr) }

//----------------------------------------

// EventDataCall fires when we call a contract, and when a contract calls another contract
type EventDataCall struct {
	TxHash     binary.HexBytes
	CallData   *CallData
	Origin     crypto.Address
	StackDepth uint64
	Return     binary.HexBytes
	Exception  *errors.Exception
}

type CallData struct {
	Caller crypto.Address
	Callee crypto.Address
	Data   binary.HexBytes
	Value  uint64
	Gas    uint64
}

// Publish/Subscribe
func PublishAccountCall(publisher event.Publisher, address crypto.Address, call *EventDataCall) error {

	return event.PublishWithEventID(publisher, EventStringAccountCall(address),
		&Event{
			Header: &Header{
				TxHash: call.TxHash,
			},
			Call: call,
		},
		map[string]interface{}{
			"address":           address,
			event.TxHashKey:     hex.EncodeUpperToString(call.TxHash),
			event.StackDepthKey: call.StackDepth,
		})
}

// Subscribe to account call event - if TxHash is provided listens for a specifc Tx otherwise captures all, if
// stackDepth is greater than or equal to 0 captures calls at a specific stack depth (useful for capturing the return
// of the root call over recursive calls
func SubscribeAccountCall(ctx context.Context, subscribable event.Subscribable, subscriber string, address crypto.Address,
	txHash []byte, stackDepth int, ch chan<- *EventDataCall) error {

	query := event.QueryForEventID(EventStringAccountCall(address))

	if len(txHash) > 0 {
		query = query.AndEquals(event.TxHashKey, hex.EncodeUpperToString(txHash))
	}

	if stackDepth >= 0 {
		query = query.AndEquals(event.StackDepthKey, stackDepth)
	}

	return event.SubscribeCallback(ctx, subscribable, subscriber, query, func(message interface{}) bool {
		ev, ok := message.(*Event)
		if ok && ev.Call != nil {
			ch <- ev.Call
		}
		return true
	})
}
