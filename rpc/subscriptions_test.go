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

package rpc

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/logging"
	"github.com/stretchr/testify/assert"
)

var mockInterval = 20 * time.Millisecond

// Test that event subscriptions can be added manually and then automatically reaped.
func TestSubReaping(t *testing.T) {
	NUM_SUBS := 100
	reaperThreshold = 200 * time.Millisecond
	reaperPeriod = 100 * time.Millisecond

	mee := event.NewEmitter(logging.NewNoopLogger())
	eSubs := NewSubscriptions(NewSubscribableService(mee, logging.NewNoopLogger()))
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

	assert.Len(t, eSubs.subs, 0)
	t.Logf("Added %d subs that were all automatically reaped.", NUM_SUBS)
}

// Test that event subscriptions can be added and removed manually.
func TestSubManualClose(t *testing.T) {
	NUM_SUBS := 100
	// Keep the reaper out of this.
	reaperThreshold = 10000 * time.Millisecond
	reaperPeriod = 10000 * time.Millisecond

	mee := event.NewEmitter(logging.NewNoopLogger())
	eSubs := NewSubscriptions(NewSubscribableService(mee, logging.NewNoopLogger()))
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
	reaperPeriod = 10000 * time.Millisecond
	// Crank it up. Now pressure is 10 times higher on each sub.
	mockInterval = 1 * time.Millisecond
	mee := event.NewEmitter(logging.NewNoopLogger())
	eSubs := NewSubscriptions(NewSubscribableService(mee, logging.NewNoopLogger()))
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
