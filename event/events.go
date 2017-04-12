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

package event

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"fmt"

	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
	go_events "github.com/tendermint/go-events"
	tm_types "github.com/tendermint/tendermint/types"
)

// TODO: [Silas] this is a compatibility layer between our event types and
// TODO: go-events. Our ultimate plan is to replace go-events with our own pub-sub
// TODO: code that will better allow us to manage and multiplex events from different
// TODO: subsystems

// Oh for a sum type
// We are using this as a marker interface for the
type anyEventData interface{}

type EventEmitter interface {
	Subscribe(subId, event string, callback func(txs.EventData)) error
	Unsubscribe(subId string) error
}

func NewEvents(eventSwitch go_events.EventSwitch, logger logging_types.InfoTraceLogger) *events {
	return &events{eventSwitch: eventSwitch, logger: logging.WithScope(logger, "Events")}
}

// Provides an EventEmitter that wraps many underlying EventEmitters as a
// convenience for Subscribing and Unsubscribing on multiple EventEmitters at
// once
func Multiplex(events ...EventEmitter) *multiplexedEvents {
	return &multiplexedEvents{events}
}

// The events struct has methods for working with events.
type events struct {
	eventSwitch go_events.EventSwitch
	logger      logging_types.InfoTraceLogger
}

// Subscribe to an event.
func (evts *events) Subscribe(subId, event string,
	callback func(txs.EventData)) error {
	cb := func(evt go_events.EventData) {
		eventData, err := mapToOurEventData(evt)
		if err != nil {
			logging.InfoMsg(evts.logger, "Failed to map go-events EventData to our EventData",
				"error", err,
				"event", event)
		}
		callback(eventData)
	}
	evts.eventSwitch.AddListenerForEvent(subId, event, cb)
	return nil
}

// Un-subscribe from an event.
func (evts *events) Unsubscribe(subId string) error {
	evts.eventSwitch.RemoveListener(subId)
	return nil
}

type multiplexedEvents struct {
	eventEmitters []EventEmitter
}

// Subscribe to an event.
func (multiEvents *multiplexedEvents) Subscribe(subId, event string,
	callback func(txs.EventData)) error {
	for _, eventEmitter := range multiEvents.eventEmitters {
		err := eventEmitter.Subscribe(subId, event, callback)
		if err != nil {
			return err
		}
	}

	return nil
}

// Un-subscribe from an event.
func (multiEvents *multiplexedEvents) Unsubscribe(subId string) error {
	for _, eventEmitter := range multiEvents.eventEmitters {
		err := eventEmitter.Unsubscribe(subId)
		if err != nil {
			return err
		}
	}

	return nil
}

// *********************************** Events ***********************************

// EventSubscribe
type EventSub struct {
	SubId string `json:"sub_id"`
}

// EventUnsubscribe
type EventUnsub struct {
	Result bool `json:"result"`
}

// EventPoll
type PollResponse struct {
	Events []interface{} `json:"events"`
}

// **************************************************************************************
// Helper function

func GenerateSubId() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("Could not generate random bytes for a subscription"+
			" id: %v", err)
	}
	rStr := hex.EncodeToString(b)
	return strings.ToUpper(rStr), nil
}

func mapToOurEventData(eventData anyEventData) (txs.EventData, error) {
	// TODO: [Silas] avoid this with a better event pub-sub system of our own
	// TODO: that maybe involves a registry of events
	switch eventData := eventData.(type) {
	case txs.EventData:
		return eventData, nil
	case tm_types.EventDataNewBlock:
		return txs.EventDataNewBlock{
			Block: eventData.Block,
		}, nil
	case tm_types.EventDataNewBlockHeader:
		return txs.EventDataNewBlockHeader{
			Header: eventData.Header,
		}, nil
	default:
		return nil, fmt.Errorf("EventData not recognised as known EventData: %v",
			eventData)
	}
}
