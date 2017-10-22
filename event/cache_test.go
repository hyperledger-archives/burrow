package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/hyperledger/burrow/logging/loggers"
)

func TestEventCache_Flush(t *testing.T) {
	evts := NewEmitter(loggers.NewNoopInfoTraceLogger())
	evts.Subscribe("nothingness", "", func(data AnyEventData) {
		// Check we are not initialising an empty buffer full of zeroed eventInfos in the Cache
		require.FailNow(t, "We should never receive a message on this switch since none are fired")
	})
	evc := NewEventCache(evts)
	evc.Flush()
	// Check after reset
	evc.Flush()
	fail := true
	pass := false
	evts.Subscribe("somethingness", "something", func(data AnyEventData) {
		if fail {
			require.FailNow(t, "Shouldn't see a message until flushed")
		}
		pass = true
	})
	evc.Fire("something", AnyEventData{})
	evc.Fire("something", AnyEventData{})
	evc.Fire("something", AnyEventData{})
	fail = false
	evc.Flush()
	assert.True(t, pass)
}

func TestEventCacheGrowth(t *testing.T) {
	evc := NewEventCache(NewEmitter(loggers.NewNoopInfoTraceLogger()))

	fireNEvents(evc, 100)
	c := cap(evc.events)
	evc.Flush()
	assert.Equal(t, c, cap(evc.events), "cache cap should remain the same after flushing events")

	fireNEvents(evc, c/maximumBufferCapacityToLengthRatio+1)
	evc.Flush()
	assert.Equal(t, c, cap(evc.events), "cache cap should remain the same after flushing more than half "+
		"the number of events as last time")

	fireNEvents(evc, c/maximumBufferCapacityToLengthRatio-1)
	evc.Flush()
	assert.True(t, c > cap(evc.events), "cache cap should drop after flushing fewer than half "+
		"the number of events as last time")

	fireNEvents(evc, c*2*maximumBufferCapacityToLengthRatio)
	evc.Flush()
	assert.True(t, c < cap(evc.events), "cache cap should grow after flushing more events than seen before")

	for numEvents := 100; numEvents >= 0; numEvents-- {
		fireNEvents(evc, numEvents)
		evc.Flush()
		assert.True(t, cap(evc.events) <= maximumBufferCapacityToLengthRatio*numEvents,
			"cap (%v) should be at most twice numEvents (%v)", cap(evc.events), numEvents)
	}
}

func fireNEvents(evc *Cache, n int) {
	for i := 0; i < n; i++ {
		evc.Fire("something", AnyEventData{})
	}
}
