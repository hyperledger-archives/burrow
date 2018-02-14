package event

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/stretchr/testify/assert"
)

func TestEventCache_Flush(t *testing.T) {
	//ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	//defer cancel()
	ctx := context.Background()
	errCh := make(chan error)
	flushed := false

	em := NewEmitter(loggers.NewNoopInfoTraceLogger())
	SubscribeCallback(ctx, em, "nothingness", NewQueryBuilder(), func(message interface{}) bool {
		// Check against sending a buffer of zeroed messages
		if message == nil {
			errCh <- fmt.Errorf("recevied empty message but none sent")
		}
		return false
	})
	evc := NewEventCache(em)
	evc.Flush()
	// Check after reset
	evc.Flush()
	SubscribeCallback(ctx, em, "somethingness", NewQueryBuilder().AndEquals("foo", "bar"),
		func(interface{}) bool {
			if flushed {
				errCh <- nil
				return true
			} else {
				errCh <- fmt.Errorf("callback was run before messages were flushed")
				return false
			}
		})

	numMessages := 3
	tags := map[string]interface{}{"foo": "bar"}
	for i := 0; i < numMessages; i++ {
		evc.Publish(ctx, fmt.Sprintf("something_%v", i), tags)
	}
	flushed = true
	evc.Flush()
	for i := 0; i < numMessages; i++ {
		select {
		case <-time.After(2 * time.Second):
			t.Fatalf("callback did not run before timeout after messages were sent")
		case err := <-errCh:
			if err != nil {
				t.Error(err)
			}
		}
	}
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
		evc.Publish(context.Background(), "something", nil)
	}
}
