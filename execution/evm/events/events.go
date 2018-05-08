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

	acm "github.com/hyperledger/burrow/account"
	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/event"
	"github.com/tmthrgd/go-hex"
)

// Functions to generate eventId strings

func EventStringAccountCall(addr acm.Address) string { return fmt.Sprintf("Acc/%s/Call", addr) }
func EventStringLogEvent(addr acm.Address) string    { return fmt.Sprintf("Log/%s", addr) }

//----------------------------------------

// EventDataCall fires when we call a contract, and when a contract calls another contract
type EventDataCall struct {
	CallData   *CallData
	Origin     acm.Address
	TxHash     []byte
	StackDepth int
	Return     []byte
	Exception  string
}

type CallData struct {
	Caller acm.Address
	Callee acm.Address
	Data   []byte
	Value  uint64
	Gas    uint64
}

// EventDataLog fires when a contract executes the LOG opcode
type EventDataLog struct {
	Address acm.Address
	Topics  []Word256
	Data    []byte
	Height  uint64
}

// Publish/Subscribe

// Subscribe to account call event - if TxHash is provided listens for a specifc Tx otherwise captures all, if
// stackDepth is greater than or equal to 0 captures calls at a specific stack depth (useful for capturing the return
// of the root call over recursive calls
func SubscribeAccountCall(ctx context.Context, subscribable event.Subscribable, subscriber string, address acm.Address,
	txHash []byte, stackDepth int, ch chan<- *EventDataCall) error {

	query := event.QueryForEventID(EventStringAccountCall(address))

	if len(txHash) > 0 {
		query = query.AndEquals(event.TxHashKey, hex.EncodeUpperToString(txHash))
	}

	if stackDepth >= 0 {
		query = query.AndEquals(event.StackDepthKey, stackDepth)
	}

	return event.SubscribeCallback(ctx, subscribable, subscriber, query, func(message interface{}) bool {
		eventDataCall, ok := message.(*EventDataCall)
		if ok {
			ch <- eventDataCall
		}
		return true
	})
}

func SubscribeLogEvent(ctx context.Context, subscribable event.Subscribable, subscriber string, address acm.Address,
	ch chan<- *EventDataLog) error {

	query := event.QueryForEventID(EventStringLogEvent(address))

	return event.SubscribeCallback(ctx, subscribable, subscriber, query, func(message interface{}) bool {
		eventDataLog, ok := message.(*EventDataLog)
		if ok {
			ch <- eventDataLog
		}
		return true
	})
}

func PublishAccountCall(publisher event.Publisher, address acm.Address, eventDataCall *EventDataCall) error {
	return event.PublishWithEventID(publisher, EventStringAccountCall(address), eventDataCall,
		map[string]interface{}{
			"address":           address,
			event.TxHashKey:     hex.EncodeUpperToString(eventDataCall.TxHash),
			event.StackDepthKey: eventDataCall.StackDepth,
		})
}

func PublishLogEvent(publisher event.Publisher, address acm.Address, eventDataLog *EventDataLog) error {
	return event.PublishWithEventID(publisher, EventStringLogEvent(address), eventDataLog,
		map[string]interface{}{"address": address})
}
