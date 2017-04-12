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
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/burrow/txs"
)

var (
	reaperTimeout   = 5 * time.Second
	reaperThreshold = 10 * time.Second
)

type EventCache struct {
	mtx    *sync.Mutex
	events []interface{}
	ts     time.Time
	subId  string
}

func newEventCache() *EventCache {
	return &EventCache{
		&sync.Mutex{},
		make([]interface{}, 0),
		time.Now(),
		"",
	}
}

func (this *EventCache) poll() []interface{} {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	var evts []interface{}
	if len(this.events) > 0 {
		evts = this.events
		this.events = []interface{}{}
	} else {
		evts = []interface{}{}
	}
	this.ts = time.Now()
	return evts
}

// Catches events that callers subscribe to and adds them to an array ready to be polled.
type EventSubscriptions struct {
	mtx          *sync.Mutex
	eventEmitter EventEmitter
	subs         map[string]*EventCache
	reap         bool
}

func NewEventSubscriptions(eventEmitter EventEmitter) *EventSubscriptions {
	es := &EventSubscriptions{
		mtx:          &sync.Mutex{},
		eventEmitter: eventEmitter,
		subs:         make(map[string]*EventCache),
		reap:         true,
	}
	go reap(es)
	return es
}

func reap(es *EventSubscriptions) {
	if !es.reap {
		return
	}
	time.Sleep(reaperTimeout)
	es.mtx.Lock()
	defer es.mtx.Unlock()
	for id, sub := range es.subs {
		if time.Since(sub.ts) > reaperThreshold {
			// Seems like Go is ok with this..
			delete(es.subs, id)
			es.eventEmitter.Unsubscribe(id)
		}
	}
	go reap(es)
}

// Add a subscription and return the generated id. Note event dispatcher
// has to call func which involves acquiring a mutex lock, so might be
// a delay - though a conflict is practically impossible, and if it does
// happen it's for an insignificant amount of time (the time it takes to
// carry out EventCache.poll() ).
func (this *EventSubscriptions) Add(eventId string) (string, error) {
	subId, errSID := GenerateSubId()
	if errSID != nil {
		return "", errSID
	}
	cache := newEventCache()
	errC := this.eventEmitter.Subscribe(subId, eventId,
		func(evt txs.EventData) {
			cache.mtx.Lock()
			defer cache.mtx.Unlock()
			cache.events = append(cache.events, evt)
		})
	cache.subId = subId
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.subs[subId] = cache
	if errC != nil {
		return "", errC
	}
	return subId, nil
}

func (this *EventSubscriptions) Poll(subId string) ([]interface{}, error) {
	sub, ok := this.subs[subId]
	if !ok {
		return nil, fmt.Errorf("Subscription not active. ID: " + subId)
	}
	return sub.poll(), nil
}

func (this *EventSubscriptions) Remove(subId string) error {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	// TODO Check this.
	_, ok := this.subs[subId]
	if !ok {
		return fmt.Errorf("Subscription not active. ID: " + subId)
	}
	delete(this.subs, subId)
	return nil
}
