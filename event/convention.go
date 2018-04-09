package event

import (
	"context"
	"fmt"
	"reflect"
)

const (
	EventIDKey     = "EventID"
	MessageTypeKey = "MessageType"
	TxTypeKey      = "TxType"
	TxHashKey      = "TxHash"
	StackDepthKey  = "StackDepth"
)

// Get a query that matches events with a specific eventID
func QueryForEventID(eventID string) *QueryBuilder {
	// Since we're accepting external output here there is a chance it won't parse...
	return NewQueryBuilder().AndEquals(EventIDKey, eventID)
}

func PublishWithEventID(publisher Publisher, eventID string, eventData interface{},
	extraTags map[string]interface{}) error {

	if extraTags[EventIDKey] != nil {
		return fmt.Errorf("PublishWithEventID was passed the extraTags with %s already set: %s = '%s'",
			EventIDKey, EventIDKey, eventID)
	}
	tags := map[string]interface{}{
		EventIDKey:     eventID,
		MessageTypeKey: reflect.TypeOf(eventData).String(),
	}
	for k, v := range extraTags {
		tags[k] = v
	}
	return publisher.Publish(context.Background(), eventData, tags)
}

// Subscribe to messages matching query and launch a goroutine to run a callback for each one. The goroutine will exit
// when the context is done or the subscription is removed.
func SubscribeCallback(ctx context.Context, subscribable Subscribable, subscriber string, query Queryable,
	callback func(message interface{}) bool) error {

	out := make(chan interface{})
	go func() {
		for msg := range out {
			if !callback(msg) {
				// Callback is requesting stop so unsubscribe and drain channel
				subscribable.Unsubscribe(context.Background(), subscriber, query)
				// Not draining channel can starve other subscribers
				for range out {
				}
				return
			}
		}
	}()
	err := subscribable.Subscribe(ctx, subscriber, query, out)
	if err != nil {
		// To clean up goroutine - otherwise subscribable should close channel for us
		close(out)
	}
	return err
}

func PublishAll(ctx context.Context, subscribable Subscribable, subscriber string, query Queryable,
	publisher Publisher, extraTags map[string]interface{}) error {

	return SubscribeCallback(ctx, subscribable, subscriber, query, func(message interface{}) bool {
		tags := make(map[string]interface{})
		for k, v := range extraTags {
			tags[k] = v
		}
		// Help! I can't tell which tags the original publisher used - so I can't forward them on
		publisher.Publish(ctx, message, tags)
		return true
	})
}
