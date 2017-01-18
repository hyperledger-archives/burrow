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

	"github.com/eris-ltd/eris-db/logging"
	"github.com/eris-ltd/eris-db/logging/loggers"
	"github.com/eris-ltd/eris-db/txs"
	go_events "github.com/tendermint/go-events"
	tm_types "github.com/tendermint/tendermint/types"
)

// TODO: [Silas] evts is a compatibility layer between our event types and
// TODO: go-events. Our ultimate plan is to replace go-events with our own pub-sub
// TODO: code that will better allow us to manage and multiplex events from different
// TODO: subsystems

// Oh for a sum type
// We are using evts as a marker interface for the
type anyEventData interface{}

type EventEmitter interface {
	Subscribe(subId, event string, callback func(txs.EventData)) error
	Unsubscribe(subId string) error
}

func NewEvents(eventSwitch *go_events.EventSwitch, logger loggers.InfoTraceLogger) *events {
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
	eventSwitch *go_events.EventSwitch
	logger      loggers.InfoTraceLogger
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
	// TODO: [Silas] avoid evts with a better event pub-sub system of our own
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
