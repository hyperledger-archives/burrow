package tendermint

import (
	"context"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/tendermint/tendermint/libs/pubsub"
	tm_types "github.com/tendermint/tendermint/types"
)

// Publishes all tendermint events available on subscribable to publisher
func PublishAllEvents(ctx context.Context, fromSubscribable event.Subscribable, subscriber string,
	toPublisher event.Publisher) error {

	var err error

	// This is a work-around for the fact we cannot access a message's tags and need a separate query for each event type
	tendermintEventTypes := []string{
		tm_types.EventBond,
		tm_types.EventCompleteProposal,
		tm_types.EventDupeout,
		tm_types.EventFork,
		tm_types.EventLock,
		tm_types.EventNewBlock,
		tm_types.EventNewBlockHeader,
		tm_types.EventNewRound,
		tm_types.EventNewRoundStep,
		tm_types.EventPolka,
		tm_types.EventRebond,
		tm_types.EventRelock,
		tm_types.EventTimeoutPropose,
		tm_types.EventTimeoutWait,
		tm_types.EventTx,
		tm_types.EventUnbond,
		tm_types.EventUnlock,
		tm_types.EventVote,
		tm_types.EventProposalHeartbeat,
	}

	for _, eventType := range tendermintEventTypes {
		publishErr := PublishEvent(ctx, fromSubscribable, subscriber, eventType, toPublisher)
		if publishErr != nil && err == nil {
			err = publishErr
		}
	}

	return err
}

func PublishEvent(ctx context.Context, fromSubscribable event.Subscribable, subscriber string, eventType string,
	toPublisher event.Publisher) error {
	tags := map[string]interface{}{
		structure.ComponentKey: "Tendermint",
		tm_types.EventTypeKey:  eventType,
		event.EventIDKey:       eventType,
	}
	return event.PublishAll(ctx, fromSubscribable, subscriber, query.WrapQuery(tm_types.QueryForEvent(eventType)),
		toPublisher, tags)
}

type eventBusSubscriber struct {
	tm_types.EventBusSubscriber
}

func EventBusAsSubscribable(eventBus tm_types.EventBusSubscriber) event.Subscribable {
	return eventBusSubscriber{eventBus}
}

func (ebs eventBusSubscriber) Subscribe(ctx context.Context, subscriber string, queryable query.Queryable,
	out chan<- interface{}) error {
	qry, err := queryable.Query()
	if err != nil {
		return err
	}
	return ebs.EventBusSubscriber.Subscribe(ctx, subscriber, qry, out)
}

func (ebs eventBusSubscriber) Unsubscribe(ctx context.Context, subscriber string, queryable query.Queryable) error {
	qry, err := queryable.Query()
	if err != nil {
		return err
	}
	return ebs.EventBusSubscriber.Unsubscribe(ctx, subscriber, qry)
}

type subscribableEventBus struct {
	event.Subscribable
}

func SubscribableAsEventBus(subscribable event.Subscribable) tm_types.EventBusSubscriber {
	return subscribableEventBus{subscribable}
}

func (seb subscribableEventBus) Subscribe(ctx context.Context, subscriber string, qry pubsub.Query,
	out chan<- interface{}) error {
	return seb.Subscribable.Subscribe(ctx, subscriber, query.WrapQuery(qry), out)
}

func (seb subscribableEventBus) Unsubscribe(ctx context.Context, subscriber string, qry pubsub.Query) error {
	return seb.Subscribable.Unsubscribe(ctx, subscriber, query.WrapQuery(qry))
}
