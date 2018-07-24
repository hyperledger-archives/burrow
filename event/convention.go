package event

import (
	"github.com/hyperledger/burrow/event/query"
)

const (
	EventTypeKey   = "EventType"
	EventIDKey     = "EventID"
	MessageTypeKey = "MessageType"
	TxHashKey      = "TxHash"
	HeightKey      = "Height"
	IndexKey       = "Index"
	StackDepthKey  = "StackDepth"
	AddressKey     = "Address"
)

// Get a query that matches events with a specific eventID
func QueryForEventID(eventID string) *query.Builder {
	// Since we're accepting external output here there is a chance it won't parse...
	return query.NewBuilder().AndEquals(EventIDKey, eventID)
}
