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
)

var (
	reaperTimeout   = 5 * time.Second
	reaperThreshold = 10 * time.Second
)

type SubscriptionsCache struct {
	mtx    *sync.Mutex
	events []interface{}
	ts     time.Time
	subId  string
}

func newSubscriptionsCache() *SubscriptionsCache {
	return &SubscriptionsCache{
		&sync.Mutex{},
		make([]interface{}, 0),
		time.Now(),
		"",
	}
}

func (subsCache *SubscriptionsCache) poll() []interface{} {
	subsCache.mtx.Lock()
	defer subsCache.mtx.Unlock()
	var evts []interface{}
	if len(subsCache.events) > 0 {
		evts = subsCache.events
		subsCache.events = []interface{}{}
	} else {
		evts = []interface{}{}
	}
	subsCache.ts = time.Now()
	return evts
}

// Catches events that callers subscribe to and adds them to an array ready to be polled.
type Subscriptions struct {
	mtx          *sync.RWMutex
	subscribable Subscribable
	subs         map[string]*SubscriptionsCache
	reap         bool
}

func NewSubscriptions(subscribable Subscribable) *Subscriptions {
	es := &Subscriptions{
		mtx:          &sync.RWMutex{},
		subscribable: subscribable,
		subs:         make(map[string]*SubscriptionsCache),
		reap:         true,
	}
	go reap(es)
	return es
}

func reap(es *Subscriptions) {
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
			es.subscribable.Unsubscribe(id)
		}
	}
	go reap(es)
}

// Add a subscription and return the generated id. Note event dispatcher
// has to call func which involves acquiring a mutex lock, so might be
// a delay - though a conflict is practically impossible, and if it does
// happen it's for an insignificant amount of time (the time it takes to
// carry out SubscriptionsCache.poll() ).
func (subs *Subscriptions) Add(eventId string) (string, error) {
	subId, errSID := GenerateSubId()
	if errSID != nil {
		return "", errSID
	}
	cache := newSubscriptionsCache()
	errC := subs.subscribable.Subscribe(subId, eventId,
		func(evt AnyEventData) {
			cache.mtx.Lock()
			defer cache.mtx.Unlock()
			cache.events = append(cache.events, evt)
		})
	cache.subId = subId
	subs.mtx.Lock()
	defer subs.mtx.Unlock()
	subs.subs[subId] = cache
	if errC != nil {
		return "", errC
	}
	return subId, nil
}

func (subs *Subscriptions) Poll(subId string) ([]interface{}, error) {
	subs.mtx.RLock()
	defer subs.mtx.RUnlock()
	sub, ok := subs.subs[subId]
	if !ok {
		return nil, fmt.Errorf("Subscription not active. ID: " + subId)
	}
	return sub.poll(), nil
}

func (subs *Subscriptions) Remove(subId string) error {
	subs.mtx.Lock()
	defer subs.mtx.Unlock()
	// TODO Check this.
	_, ok := subs.subs[subId]
	if !ok {
		return fmt.Errorf("Subscription not active. ID: " + subId)
	}
	delete(subs.subs, subId)
	return nil
}
