package event

import (
	"context"

	"github.com/hyperledger/burrow/event/query"
)

const (
	EventTypeKey   = "EventType"
	EventIDKey     = "EventID"
	MessageTypeKey = "MessageType"
	TxTypeKey      = "TxType"
	TxHashKey      = "TxHash"
	HeightKey      = "Height"
	IndexKey       = "Index"
	NameKey        = "Name"
	PermissionKey  = "Permission"
	StackDepthKey  = "StackDepth"
	AddressKey     = "Address"
	OriginKey      = "Origin"
	CalleeKey      = "Callee"
	CallerKey      = "Caller"
	ValueKey       = "Value"
	GasKey         = "Gas"
	ExceptionKey   = "Exception"
)

// Get a query that matches events with a specific eventID
func QueryForEventID(eventID string) *query.Builder {
	// Since we're accepting external output here there is a chance it won't parse...
	return query.NewBuilder().AndEquals(EventIDKey, eventID)
}

// Subscribe to messages matching query and launch a goroutine to run a callback for each one. The goroutine will exit
// when the context is done or the subscription is removed.
func SubscribeCallback(ctx context.Context, subscribable Subscribable, subscriber string, queryable query.Queryable,
	callback func(message interface{}) bool) error {

	out := make(chan interface{})
	go func() {
		for msg := range out {
			if !callback(msg) {
				// Callback is requesting stop so unsubscribe and drain channel
				subscribable.Unsubscribe(context.Background(), subscriber, queryable)
				// Not draining channel can starve other subscribers
				for range out {
				}
				return
			}
		}
	}()
	err := subscribable.Subscribe(ctx, subscriber, queryable, out)
	if err != nil {
		// To clean up goroutine - otherwise subscribable should close channel for us
		close(out)
	}
	return err
}

func PublishAll(ctx context.Context, subscribable Subscribable, subscriber string, queryable query.Queryable,
	publisher Publisher, extraTags map[string]interface{}) error {

	return SubscribeCallback(ctx, subscribable, subscriber, queryable, func(message interface{}) bool {
		tags := make(map[string]interface{})
		for k, v := range extraTags {
			tags[k] = v
		}
		// Help! I can't tell which tags the original publisher used - so I can't forward them on
		publisher.Publish(ctx, message, TagMap(tags))
		return true
	})
}
