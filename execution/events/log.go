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

	"strings"

	"github.com/hyperledger/burrow/binary"
	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/txs"
	"github.com/tmthrgd/go-hex"
)

// Functions to generate eventId strings

func EventStringLogEvent(addr crypto.Address) string { return fmt.Sprintf("Log/%s", addr) }

//----------------------------------------

// EventDataLog fires when a contract executes the LOG opcode
type EventDataLog struct {
	Height  uint64
	Address crypto.Address
	Topics  []Word256
	Data    binary.HexBytes
}

// Publish/Subscribe
func PublishLogEvent(publisher event.Publisher, tx *txs.Tx, log *EventDataLog) error {
	ev := &Event{
		Header: &Header{
			TxType:    tx.Type(),
			TxHash:    tx.Hash(),
			EventType: TypeLog,
			EventID:   EventStringLogEvent(log.Address),
			Height:    log.Height,
		},
		Log: log,
	}
	return publisher.Publish(context.Background(), ev, ev.Tags())
}

func SubscribeLogEvent(ctx context.Context, subscribable event.Subscribable, subscriber string, address crypto.Address,
	ch chan<- *EventDataLog) error {

	qry := event.QueryForEventID(EventStringLogEvent(address))

	return event.SubscribeCallback(ctx, subscribable, subscriber, qry, func(message interface{}) (stop bool) {
		ev, ok := message.(*Event)
		if ok && ev.Log != nil {
			ch <- ev.Log
		}
		return
	})
}

// Tags
const logNTextTopicCutset = "\x00"

var logTagKeys []string
var logNTopicIndex = make(map[string]int, 5)
var logNTextTopicIndex = make(map[string]int, 5)

func init() {
	for i := 0; i <= 4; i++ {
		logN := event.LogNKey(i)
		logTagKeys = append(logTagKeys, event.LogNKey(i))
		logNText := event.LogNTextKey(i)
		logTagKeys = append(logTagKeys, logNText)
		logNTopicIndex[logN] = i
		logNTextTopicIndex[logNText] = i
	}
	logTagKeys = append(logTagKeys, event.AddressKey)
}

func (log *EventDataLog) Get(key string) (string, bool) {
	var value interface{}
	switch key {
	case event.AddressKey:
		value = log.Address
	default:
		if i, ok := logNTopicIndex[key]; ok {
			return hex.EncodeUpperToString(log.GetTopic(i).Bytes()), true
		}
		if i, ok := logNTextTopicIndex[key]; ok {
			return strings.Trim(string(log.GetTopic(i).Bytes()), logNTextTopicCutset), true
		}
		return "", false
	}
	return query.StringFromValue(value), true
}

func (log *EventDataLog) GetTopic(i int) Word256 {
	if i < len(log.Topics) {
		return log.Topics[i]
	}
	return Word256{}
}

func (log *EventDataLog) Len() int {
	return len(logTagKeys)
}

func (log *EventDataLog) Keys() []string {
	return logTagKeys
}
