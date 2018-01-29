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
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/tendermint/tmlibs/common"
	go_events "github.com/tendermint/tmlibs/events"
)

type Subscribable interface {
	Subscribe(subId, event string, callback func(AnyEventData)) error
	Unsubscribe(subId string) error
}

type Fireable interface {
	Fire(event string, data interface{})
}

type Emitter interface {
	Fireable
	go_events.EventSwitch
	Subscribable
}

// The events struct has methods for working with events.
type emitter struct {
	// Bah, Service infects everything
	*common.BaseService
	eventSwitch go_events.EventSwitch
	logger      logging_types.InfoTraceLogger
}

var _ Emitter = &emitter{}

func NewEmitter(logger logging_types.InfoTraceLogger) *emitter {
	return WrapEventSwitch(go_events.NewEventSwitch(), logger)
}

func WrapEventSwitch(eventSwitch go_events.EventSwitch, logger logging_types.InfoTraceLogger) *emitter {
	eventSwitch.Start()
	return &emitter{
		BaseService: common.NewBaseService(nil, "BurrowEventEmitter", eventSwitch),
		eventSwitch: eventSwitch,
		logger:      logger.With(structure.ComponentKey, "Events"),
	}
}

// Fireable
func (evts *emitter) Fire(event string, eventData interface{}) {
	evts.eventSwitch.FireEvent(event, eventData)
}

func (evts *emitter) FireEvent(event string, data go_events.EventData) {
	evts.Fire(event, data)
}

// EventSwitch
func (evts *emitter) AddListenerForEvent(listenerID, event string, cb go_events.EventCallback) {
	evts.eventSwitch.AddListenerForEvent(listenerID, event, cb)
}

func (evts *emitter) RemoveListenerForEvent(event string, listenerID string) {
	evts.eventSwitch.RemoveListenerForEvent(event, listenerID)
}

func (evts *emitter) RemoveListener(listenerID string) {
	evts.eventSwitch.RemoveListener(listenerID)
}

// Subscribe to an event.
func (evts *emitter) Subscribe(subId, event string, callback func(AnyEventData)) error {
	logging.TraceMsg(evts.logger, "Subscribing to event",
		structure.ScopeKey, "events.Subscribe", "subId", subId, "event", event)
	evts.eventSwitch.AddListenerForEvent(subId, event, func(eventData go_events.EventData) {
		if eventData == nil {
			logging.TraceMsg(evts.logger, "Sent nil go-events EventData")
			return
		}
		callback(MapToAnyEventData(eventData))
	})
	return nil
}

// Un-subscribe from an event.
func (evts *emitter) Unsubscribe(subId string) error {
	logging.TraceMsg(evts.logger, "Unsubscribing from event",
		structure.ScopeKey, "events.Unsubscribe", "subId", subId)
	evts.eventSwitch.RemoveListener(subId)
	return nil
}

// Provides an Emitter that wraps many underlying EventEmitters as a
// convenience for Subscribing and Unsubscribing on multiple EventEmitters at
// once
func Multiplex(events ...Emitter) *multiplexedEvents {
	return &multiplexedEvents{
		BaseService:   common.NewBaseService(nil, "BurrowMultiplexedEventEmitter", nil),
		eventEmitters: events,
	}
}

type multiplexedEvents struct {
	*common.BaseService
	eventEmitters []Emitter
}

var _ Emitter = &multiplexedEvents{}

// Subscribe to an event.
func (multiEvents *multiplexedEvents) Subscribe(subId, event string, cb func(AnyEventData)) error {
	for _, evts := range multiEvents.eventEmitters {
		err := evts.Subscribe(subId, event, cb)
		if err != nil {
			return err
		}
	}
	return nil
}

func (multiEvents *multiplexedEvents) Unsubscribe(subId string) error {
	for _, evts := range multiEvents.eventEmitters {
		err := evts.Unsubscribe(subId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (multiEvents *multiplexedEvents) Fire(event string, eventData interface{}) {
	for _, evts := range multiEvents.eventEmitters {
		evts.Fire(event, eventData)
	}
}

func (multiEvents *multiplexedEvents) FireEvent(event string, eventData go_events.EventData) {
	multiEvents.Fire(event, eventData)
}

// EventSwitch
func (multiEvents *multiplexedEvents) AddListenerForEvent(listenerID, event string, cb go_events.EventCallback) {
	for _, evts := range multiEvents.eventEmitters {
		evts.AddListenerForEvent(listenerID, event, cb)
	}
}

func (multiEvents *multiplexedEvents) RemoveListenerForEvent(event string, listenerID string) {
	for _, evts := range multiEvents.eventEmitters {
		evts.RemoveListenerForEvent(event, listenerID)
	}
}

func (multiEvents *multiplexedEvents) RemoveListener(listenerID string) {
	for _, evts := range multiEvents.eventEmitters {
		evts.RemoveListener(listenerID)
	}
}

type noOpFireable struct {
}

func (*noOpFireable) Fire(string, interface{}) {

}

func NewNoOpFireable() Fireable {
	return &noOpFireable{}
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
		return "", fmt.Errorf("could not generate random bytes for a subscription"+
			" id: %v", err)
	}
	rStr := hex.EncodeToString(b)
	return strings.ToUpper(rStr), nil
}
