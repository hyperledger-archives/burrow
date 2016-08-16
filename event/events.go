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

package event

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"fmt"

	"github.com/eris-ltd/eris-db/txs"
	go_events "github.com/tendermint/go-events"
	tm_types "github.com/tendermint/tendermint/types"
)

// TODO improve
// TODO: [ben] yes please ^^^
// [ben] To improve this we will switch out go-events with eris-db/event so
// that there is no need anymore for this poor wrapper.

// Oh for a sum type
type anyEventData interface{}

type EventEmitter interface {
	Subscribe(subId, event string, callback func(txs.EventData)) error
	Unsubscribe(subId string) error
}

func NewEvents(eventSwitch *go_events.EventSwitch) *events {
	return &events{eventSwitch}
}

// Provides an EventEmitter that wraps many underlying EventEmitters as a
// convenience for Subscribing and Unsubscribing on multiple EventEmitters at
// once
func Multiplex(events ...EventEmitter) *multiplexedEvents {
	return &multiplexedEvents{events}
}

// The events struct has methods for working with events.
type events struct {
	eventSwitch *go_events.EventSwitch
}

// Subscribe to an event.
func (this *events) Subscribe(subId, event string,
	callback func(txs.EventData)) error {
	cb := func(evt go_events.EventData) {
		eventData, err := mapToOurEventData(evt)
		if err != nil {
			panic(fmt.Sprintf("Failed to map go-events EventData to our EventData %v",
				err))
		}
		callback(eventData)
	}
	this.eventSwitch.AddListenerForEvent(subId, event, cb)
	return nil
}

// Un-subscribe from an event.
func (this *events) Unsubscribe(subId string) error {
	this.eventSwitch.RemoveListener(subId)
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
		return "", err
	}
	rStr := hex.EncodeToString(b)
	return strings.ToUpper(rStr), nil
}

func mapToOurEventData(eventData anyEventData) (txs.EventData, error) {
	// While we depend on go-events in the way we do, we don't have much choice
	// than to use a generic interface like anyEventData with a type switch.

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
