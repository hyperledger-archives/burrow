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
	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
)

// Functions to generate eventId strings

func EventStringLogEvent(addr crypto.Address) string { return fmt.Sprintf("Log/%s", addr) }

//----------------------------------------

// EventDataLog fires when a contract executes the LOG opcode
type EventDataLog struct {
	TxHash  binary.HexBytes
	Address crypto.Address
	Topics  []Word256
	Data    binary.HexBytes
	Height  uint64
}

// Publish/Subscribe
func PublishLogEvent(publisher event.Publisher, address crypto.Address, log *EventDataLog) error {

	return event.PublishWithEventID(publisher, EventStringLogEvent(address),
		&Event{
			Header: &Header{
				TxHash: log.TxHash,
			},
			Log: log,
		},
		map[string]interface{}{"address": address})
}

func SubscribeLogEvent(ctx context.Context, subscribable event.Subscribable, subscriber string, address crypto.Address,
	ch chan<- *EventDataLog) error {

	query := event.QueryForEventID(EventStringLogEvent(address))

	return event.SubscribeCallback(ctx, subscribable, subscriber, query, func(message interface{}) bool {
		ev, ok := message.(*Event)
		if ok && ev.Log != nil {
			ch <- ev.Log
		}
		return true
	})
}
