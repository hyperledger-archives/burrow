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
	"testing"

	"sync"
	"time"

	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
)

func TestMultiplexedEvents(t *testing.T) {
	emitter1 := newMockEventEmitter()
	emitter2 := newMockEventEmitter()
	emitter12 := Multiplex(emitter1, emitter2)

	eventData1 := make(map[txs.EventData]int)
	eventData2 := make(map[txs.EventData]int)
	eventData12 := make(map[txs.EventData]int)

	mutex1 := &sync.Mutex{}
	mutex2 := &sync.Mutex{}
	mutex12 := &sync.Mutex{}

	emitter12.Subscribe("Sub12", "Event12", func(eventData txs.EventData) {
		mutex12.Lock()
		eventData12[eventData] = 1
		mutex12.Unlock()
	})
	emitter1.Subscribe("Sub1", "Event1", func(eventData txs.EventData) {
		mutex1.Lock()
		eventData1[eventData] = 1
		mutex1.Unlock()
	})
	emitter2.Subscribe("Sub2", "Event2", func(eventData txs.EventData) {
		mutex2.Lock()
		eventData2[eventData] = 1
		mutex2.Unlock()
	})

	time.Sleep(2 * mockInterval)

	err := emitter12.Unsubscribe("Sub12")
	assert.NoError(t, err)
	err = emitter1.Unsubscribe("Sub2")
	assert.NoError(t, err)
	err = emitter2.Unsubscribe("Sub2")
	assert.NoError(t, err)

	mutex1.Lock()
	defer mutex1.Unlock()
	mutex2.Lock()
	defer mutex2.Unlock()
	mutex12.Lock()
	defer mutex12.Unlock()
	assert.Equal(t, map[txs.EventData]int{mockEventData{"Sub1", "Event1"}: 1},
		eventData1)
	assert.Equal(t, map[txs.EventData]int{mockEventData{"Sub2", "Event2"}: 1},
		eventData2)
	assert.Equal(t, map[txs.EventData]int{mockEventData{"Sub12", "Event12"}: 1},
		eventData12)

}
