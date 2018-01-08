package event

// When exceeded we will trim the buffer's backing array capacity to avoid excessive
// allocation

const maximumBufferCapacityToLengthRatio = 2

// An Cache buffers events for a Fireable
// All events are cached. Filtering happens on Flush
type Cache struct {
	evsw   Fireable
	events []eventInfo
}

var _ Fireable = &Cache{}

// Create a new Cache with an EventSwitch as backend
func NewEventCache(evsw Fireable) *Cache {
	return &Cache{
		evsw: evsw,
	}
}

// a cached event
type eventInfo struct {
	event string
	data  AnyEventData
}

// Cache an event to be fired upon finality.
func (evc *Cache) Fire(event string, eventData interface{}) {
	// append to list (go will grow our backing array exponentially)
	evc.events = append(evc.events, eventInfo{event: event, data: MapToAnyEventData(eventData)})
}

// Fire events by running evsw.Fire on all cached events. Blocks.
// Clears cached events
func (evc *Cache) Flush() {
	for _, ei := range evc.events {
		evc.evsw.Fire(ei.event, ei.data)
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
}
