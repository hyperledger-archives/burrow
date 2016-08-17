package event

import (
	"testing"

	"sync"
	"time"

	"github.com/stretchr/testify/assert"
	evts "github.com/tendermint/go-events"
	"github.com/eris-ltd/eris-db/txs"
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

	time.Sleep(mockInterval)

	allEventData := make(map[txs.EventData]int)
	for k, v := range eventData1 {
		allEventData[k] = v
	}
	for k, v := range eventData2 {
		allEventData[k] = v
	}

	assert.Equal(t, map[txs.EventData]int{mockEventData{"Sub1", "Event1"}: 1},
		eventData1)
	assert.Equal(t, map[txs.EventData]int{mockEventData{"Sub2", "Event2"}: 1},
		eventData2)
	assert.Equal(t, map[txs.EventData]int{mockEventData{"Sub12", "Event12"}: 1},
		eventData12)

	assert.NotEmpty(t, allEventData, "Some events should have been published")
}
