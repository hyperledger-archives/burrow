package event

import (
	"context"
)

// When exceeded we will trim the buffer's backing array capacity to avoid excessive
// allocation
const maximumBufferCapacityToLengthRatio = 2

// A Cache buffers events for a Publisher.
type Cache struct {
	events []messageInfo
}

// If message implement this interface we will provide them with an index in the cache
type Indexable interface {
	ProvideIndex(index uint64)
}

var _ Publisher = &Cache{}

// Create a new Cache with an EventSwitch as backend
func NewEventCache() *Cache {
	return &Cache{}
}

// a cached event
type messageInfo struct {
	// Hmm... might be unintended interactions with pushing a deadline into a cache - though usually we publish with an
	// empty context
	ctx     context.Context
	message interface{}
	tags    Tags
}

// Cache an event to be fired upon finality.
func (evc *Cache) Publish(ctx context.Context, message interface{}, tags Tags) error {
	// append to list (go will grow our backing array exponentially)
	evc.events = append(evc.events, messageInfo{
		ctx:     ctx,
		message: evc.provideIndex(message),
		tags:    tags,
	})
	return nil
}

func (evc *Cache) Flush(publisher Publisher) error {
	err := evc.Sync(publisher)
	if err != nil {
		return err
	}
	evc.Reset()
	return nil
}

// Clears cached events by flushing them to Publisher
func (evc *Cache) Sync(publisher Publisher) error {
	var err error
	for _, mi := range evc.events {
		publishErr := publisher.Publish(mi.ctx, mi.message, mi.tags)
		// Capture first by try to sync the rest
		if publishErr != nil && err == nil {
			err = publishErr
		}
	}
	return err
}

func (evc *Cache) Reset() {
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
}

func (evc *Cache) provideIndex(message interface{}) interface{} {
	if im, ok := message.(Indexable); ok {
		im.ProvideIndex(uint64(len(evc.events)))
	}
	return message
}
