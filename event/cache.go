package event

import (
	"context"
)

// When exceeded we will trim the buffer's backing array capacity to avoid excessive
// allocation
const maximumBufferCapacityToLengthRatio = 2

// A Cache buffers events for a Publisher.
type Cache struct {
	publisher Publisher
	events    []messageInfo
}

var _ Publisher = &Cache{}

// Create a new Cache with an EventSwitch as backend
func NewEventCache(publisher Publisher) *Cache {
	return &Cache{
		publisher: publisher,
	}
}

// a cached event
type messageInfo struct {
	// Hmm... might be unintended interactions with pushing a deadline into a cache - though usually we publish with an
	// empty context
	ctx     context.Context
	message interface{}
	tags    map[string]interface{}
}

// Cache an event to be fired upon finality.
func (evc *Cache) Publish(ctx context.Context, message interface{}, tags map[string]interface{}) error {
	// append to list (go will grow our backing array exponentially)
	evc.events = append(evc.events, messageInfo{
		ctx:     ctx,
		message: message,
		tags:    tags,
	})
	return nil
}

// Clears cached events by flushing them to Publisher
func (evc *Cache) Flush() error {
	var err error
	for _, mi := range evc.events {
		publishErr := evc.publisher.Publish(mi.ctx, mi.message, mi.tags)
		// Capture first by try to flush the rest
		if publishErr != nil && err == nil {
			err = publishErr
		}
	}
	// Clear the buffer by re-slicing its length to zero
	if cap(evc.events) > len(evc.events)*maximumBufferCapacityToLengthRatio {
		// Trim the backing array capacity when it is more than double the length of the slice to avoid tying up memory
		// after a spike in the number of events to buffer
		evc.events = evc.events[:0:len(evc.events)]
	} else {
		// Re-slice the length to 0 to clear buffer but hang on to spare capacity in backing array that has been added
		// in previous cache round
		evc.events = evc.events[:0]
	}
	return err
}
