package pipe

import (
	evts "github.com/tendermint/tendermint/events"
)

// TODO improve

// The events struct has methods for working with events.
type events struct {
	eventSwitch *evts.EventSwitch
}

func newEvents(eventSwitch *evts.EventSwitch) *events {
	return &events{eventSwitch}
}

// Subscribe to an event.
func (this *events) Subscribe(subId, event string, callback func(interface{})) (bool, error) {
	this.eventSwitch.AddListenerForEvent(subId, event, callback)
	return true, nil
}

// Un-subscribe from an event.
func (this *events) Unsubscribe(subId string) (bool, error) {
	this.eventSwitch.RemoveListener(subId)
	return true, nil
}
