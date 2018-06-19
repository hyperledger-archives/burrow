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
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/txs"
	"github.com/tmthrgd/go-hex"
)

// Functions to generate eventId strings

func EventStringAccountCall(addr crypto.Address) string { return fmt.Sprintf("Acc/%s/Call", addr) }

//----------------------------------------

// EventDataCall fires when we call a contract, and when a contract calls another contract
type EventDataCall struct {
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

var callTagKeys = []string{
	event.CalleeKey,
	event.CallerKey,
	event.ValueKey,
	event.GasKey,
	event.StackDepthKey,
	event.OriginKey,
	event.ExceptionKey,
}

// Implements Tags for events
func (call *EventDataCall) Get(key string) (string, bool) {
	var value interface{}
	switch key {
	case event.CalleeKey:
		value = call.CallData.Callee
	case event.CallerKey:
		value = call.CallData.Caller
	case event.ValueKey:
		value = call.CallData.Value
	case event.GasKey:
		value = call.CallData.Gas
	case event.StackDepthKey:
		value = call.StackDepth
	case event.OriginKey:
		value = call.Origin
	case event.ExceptionKey:
		value = call.Exception
	default:
		return "", false
	}
	return query.StringFromValue(value), true
}

func (call *EventDataCall) Len() int {
	return len(callTagKeys)
}

func (call *EventDataCall) Map() map[string]interface{} {
	tags := make(map[string]interface{})
	for _, key := range callTagKeys {
		tags[key], _ = call.Get(key)
	}
	return tags
}

func (call *EventDataCall) Keys() []string {
	return callTagKeys
}
func (call *EventDataCall) Tags(tags map[string]interface{}) map[string]interface{} {
	tags[event.CalleeKey] = call.CallData.Callee
	tags[event.CallerKey] = call.CallData.Caller
	tags[event.ValueKey] = call.CallData.Value
	tags[event.GasKey] = call.CallData.Gas
	tags[event.StackDepthKey] = call.StackDepth
	tags[event.OriginKey] = call.Origin
	if call.Exception != nil {
		tags[event.ExceptionKey] = call.Exception
	}
	return tags
}

// Publish/Subscribe
func PublishAccountCall(publisher event.Publisher, tx *txs.Tx, height uint64, call *EventDataCall) error {
	eventID := EventStringAccountCall(call.CallData.Callee)
	ev := &Event{
		Header: &Header{
			TxType:    tx.Type(),
			TxHash:    tx.Hash(),
			EventType: TypeCall,
			EventID:   eventID,
			Height:    height,
		},
		Call: call,
	}
	return publisher.Publish(context.Background(), ev, ev.Tags())
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
