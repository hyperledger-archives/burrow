package event

import go_events "github.com/tendermint/tmlibs/events"

// An EventCache buffers events for a Fireable
type EventCache struct {
	fireable go_events.Fireable
	events   []eventInfo
}

func NewEventCache(fireable go_events.Fireable) *EventCache {
	return &EventCache{
		fireable: fireable,
	}
}

type eventInfo struct {
	event string
	data  AnyEventData
}

// Cache an event to be fired upon finality.
func (evc *EventCache) FireEvent(event string, eventData go_events.EventData) {
	// append to list
	evc.events = append(evc.events, eventInfo{event: event, data: MapToAnyEventData(eventData)})
}

// Fire events by running fireable.FireEvent on all cached events. Blocks.
// Clears cached events
func (evc *EventCache) Flush() {
	for _, ei := range evc.events {
		evc.fireable.FireEvent(ei.event, ei.data)
	}
	evc.events = nil
}
