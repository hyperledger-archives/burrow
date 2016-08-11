package event

import (
	"encoding/hex"
	"fmt"
	"runtime"
	"testing"
	"time"

	"sync"

	"github.com/stretchr/testify/assert"
	evts "github.com/tendermint/go-events"
)

var mockInterval = 10 * time.Millisecond

type mockSub struct {
	subId    string
	eventId  string
	f        func(evts.EventData)
	shutdown bool
	sdChan   chan struct{}
}

type mockEventData struct {
	subId   string
	eventId string
}

// A mock event
func newMockSub(subId, eventId string, f func(evts.EventData)) mockSub {
	return mockSub{subId, eventId, f, false, make(chan struct{})}
}

type mockEventEmitter struct {
	subs  map[string]mockSub
	mutex *sync.Mutex
}

func newMockEventEmitter() *mockEventEmitter {
	return &mockEventEmitter{make(map[string]mockSub), &sync.Mutex{}}
}

func (this *mockEventEmitter) Subscribe(subId, eventId string, callback func(evts.EventData)) error {
	if _, ok := this.subs[subId]; ok {
		return nil
	}
	me := newMockSub(subId, eventId, callback)
	this.mutex.Lock()
	this.subs[subId] = me
	this.mutex.Unlock()

	go func() {
		<-me.sdChan
		me.shutdown = true
	}()
	go func() {
		for {
			if !me.shutdown {
				me.f(mockEventData{subId, eventId})
			} else {
				this.mutex.Lock()
				delete(this.subs, subId)
				this.mutex.Unlock()
				return
			}
			time.Sleep(mockInterval)
		}
	}()
	return nil
}

func (this *mockEventEmitter) Unsubscribe(subId string) error {
	sub, ok := this.subs[subId]
	if !ok {
		return nil
	}
	sub.shutdown = true
	delete(this.subs, subId)
	return nil
}

// Test that event subscriptions can be added manually and then automatically reaped.
func TestSubReaping(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	NUM_SUBS := 100
	reaperThreshold = 200 * time.Millisecond
	reaperTimeout = 100 * time.Millisecond

	mee := newMockEventEmitter()
	eSubs := NewEventSubscriptions(mee)
	doneChan := make(chan error)
	go func() {
		for i := 0; i < NUM_SUBS; i++ {
			time.Sleep(2 * time.Millisecond)
			go func() {
				id, err := eSubs.Add("WeirdEvent")
				if err != nil {
					doneChan <- err
					return
				}
				if len(id) != 64 {
					doneChan <- fmt.Errorf("Id not of length 64")
					return
				}
				_, err2 := hex.DecodeString(id)
				if err2 != nil {
					doneChan <- err2
				}

				doneChan <- nil
			}()
		}
	}()
	k := 0
	for k < NUM_SUBS {
		err := <-doneChan
		assert.NoError(t, err)
		k++
	}
	time.Sleep(1100 * time.Millisecond)

	assert.Len(t, mee.subs, 0)
	assert.Len(t, eSubs.subs, 0)
	t.Logf("Added %d subs that were all automatically reaped.", NUM_SUBS)
}

// Test that event subscriptions can be added and removed manually.
func TestSubManualClose(t *testing.T) {
	NUM_SUBS := 100
	// Keep the reaper out of this.
	reaperThreshold = 10000 * time.Millisecond
	reaperTimeout = 10000 * time.Millisecond

	mee := newMockEventEmitter()
	eSubs := NewEventSubscriptions(mee)
	doneChan := make(chan error)
	go func() {
		for i := 0; i < NUM_SUBS; i++ {
			time.Sleep(2 * time.Millisecond)
			go func() {
				id, err := eSubs.Add("WeirdEvent")
				if err != nil {
					doneChan <- err
					return
				}
				if len(id) != 64 {
					doneChan <- fmt.Errorf("Id not of length 64")
					return
				}
				_, err2 := hex.DecodeString(id)
				if err2 != nil {
					doneChan <- err2
				}
				time.Sleep(100 * time.Millisecond)
				err3 := eSubs.Remove(id)
				if err3 != nil {
					doneChan <- err3
				}
				doneChan <- nil
			}()
		}
	}()
	k := 0
	for k < NUM_SUBS {
		err := <-doneChan
		assert.NoError(t, err)
		k++
	}

	assert.Len(t, eSubs.subs, 0)
	t.Logf("Added %d subs that were all closed down by unsubscribing.", NUM_SUBS)
}

// Test that the system doesn't fail under high pressure.
func TestSubFlooding(t *testing.T) {
	NUM_SUBS := 100
	// Keep the reaper out of this.
	reaperThreshold = 10000 * time.Millisecond
	reaperTimeout = 10000 * time.Millisecond
	// Crank it up. Now pressure is 10 times higher on each sub.
	mockInterval = 1 * time.Millisecond
	mee := newMockEventEmitter()
	eSubs := NewEventSubscriptions(mee)
	doneChan := make(chan error)
	go func() {
		for i := 0; i < NUM_SUBS; i++ {
			time.Sleep(1 * time.Millisecond)
			go func() {
				id, err := eSubs.Add("WeirdEvent")
				if err != nil {
					doneChan <- err
					return
				}
				if len(id) != 64 {
					doneChan <- fmt.Errorf("Id not of length 64")
					return
				}
				_, err2 := hex.DecodeString(id)
				if err2 != nil {
					doneChan <- err2
				}
				time.Sleep(1000 * time.Millisecond)
				err3 := eSubs.Remove(id)
				if err3 != nil {
					doneChan <- err3
				}
				doneChan <- nil
			}()
		}
	}()
	k := 0
	for k < NUM_SUBS {
		err := <-doneChan
		assert.NoError(t, err)
		k++
	}

	assert.Len(t, eSubs.subs, 0)
	t.Logf("Added %d subs that all received 1000 events each. They were all closed down by unsubscribing.", NUM_SUBS)
}
