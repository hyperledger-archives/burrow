package erisdb

import (
	"fmt"
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	"sync"
	"time"
)

const (
	REAPER_TIMEOUT   = 5 * time.Second
	REAPER_THRESHOLD = 10 * time.Second
)

type EventCache struct {
	mtx    *sync.Mutex
	events []interface{}
	ts     time.Time
	subId string
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
		evts =  []interface{}{}
	}
	this.ts = time.Now()
	return evts
}

// Catches events that callers subscribe to and adds them to an array ready to be polled.
type EventSubscriptions struct {
	mtx  *sync.Mutex
	pipe ep.Pipe
	subs map[string]*EventCache
	reap bool
}

func NewEventSubscriptions(pipe ep.Pipe) *EventSubscriptions {
	es := &EventSubscriptions{
		mtx:  &sync.Mutex{},
		pipe: pipe,
		subs: make(map[string]*EventCache),
		reap: true,
	}
	go reap(es)
	return es
}

func reap(es *EventSubscriptions) {
	if !es.reap {
		return
	}
	time.Sleep(REAPER_TIMEOUT)
	es.mtx.Lock()
	defer es.mtx.Unlock()
	for id, sub := range es.subs {
		if time.Since(sub.ts) > REAPER_THRESHOLD {
			fmt.Println("[SUBSCRIPTION REAPER] Reaping sub: " + sub.subId)
			// Seems like Go is ok with this..
			delete(es.subs, id)
			es.pipe.Events().Unsubscribe(id)
		}
	}
	go reap(es)
}

// Add a subscription and return the generated id. Note event dispatcher
// has to call func which involves aquiring a mutex lock, so might be
// a delay - though a conflict is practically impossible, and if it does
// happen it's for an insignificant amount of time (the time it takes to
// carry out EventCache.poll() ).
func (this *EventSubscriptions) add(eventId string) (string, error) {
	subId, errSID := generateSubId()
	if errSID != nil {
		return "", errSID
	}
	cache := newEventCache()
	_, errC := this.pipe.Events().Subscribe(subId, eventId,
		func(evt interface{}) {
			cache.mtx.Lock()
			defer cache.mtx.Unlock()
			cache.events = append(cache.events, evt)
			cache.ts = time.Now()
		})
	cache.subId = subId
	this.subs[subId] = cache
	if errC != nil {
		return "", errC
	}
	return subId, nil
}

func (this *EventSubscriptions) poll(subId string) ([]interface{}, error) {
	sub, ok := this.subs[subId]
	if !ok {
		return nil, fmt.Errorf("Subscription not active. ID: " + subId)
	}
	return sub.poll(), nil
}

func (this *EventSubscriptions) remove(subId string) error {
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
