package rpcevents

import (
	"fmt"

	"context"

	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/events/pbevents"
	"github.com/tendermint/tendermint/libs/pubsub"
)

type executionEventsServer struct {
	eventsProvider events.Provider
	emitter        event.Emitter
	tip            bcm.TipInfo
}

func NewExecutionEventsServer(eventsProvider events.Provider, emitter event.Emitter,
	tip bcm.TipInfo) pbevents.ExecutionEventsServer {

	return &executionEventsServer{
		eventsProvider: eventsProvider,
		emitter:        emitter,
		tip:            tip,
	}
}

func (ees *executionEventsServer) GetEvents(request *pbevents.GetEventsRequest,
	stream pbevents.ExecutionEvents_GetEventsServer) error {

	blockRange := request.GetBlockRange()
	start, end, streaming := blockRange.Bounds(ees.tip.LastBlockHeight())
	qry, err := query.NewBuilder(request.Query).Query()
	if err != nil {
		return fmt.Errorf("could not parse event query: %v", err)
	}

	if !streaming {
		return ees.steamEvents(stream, start, end, 1, qry)
	}

	// Streaming
	if err != nil {
		return err
	}

	out, err := tendermint.SubscribeNewBlock(context.Background(), ees.emitter)
	if err != nil {
		return err
	}

	for newBlock := range out {
		if newBlock == nil {
			return fmt.Errorf("received non-new-block event when subscribed with query")
		}
		if newBlock.Block == nil {
			return fmt.Errorf("new block contains no block info: %v", newBlock)
		}
		height := uint64(newBlock.Block.Height)
		start = end
		end = events.NewKey(height, 0)
		err := ees.steamEvents(stream, start, end, 1, qry)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ees *executionEventsServer) steamEvents(stream pbevents.ExecutionEvents_GetEventsServer, start, end events.Key,
	batchSize uint64, qry pubsub.Query) error {

	var streamErr error
	buf := new(pbevents.GetEventsResponse)

	batchStart := start.Height()
	_, err := ees.eventsProvider.GetEvents(start, end, func(ev *events.Event) (stop bool) {
		if qry.Matches(ev) {
			// Start a new batch, flush the last lot
			if ev.Header.Index == 0 && (ev.Header.Height-batchStart)%batchSize == 0 {
				streamErr = flush(stream, buf)
				if streamErr != nil {
					return true
				}
				batchStart = ev.Header.Height
				buf = new(pbevents.GetEventsResponse)
			}
			buf.Events = append(buf.Events, pbevents.GetExecutionEvent(ev))
		}
		return false
	})
	if err != nil {
		return err
	}
	if streamErr != nil {
		return streamErr
	}
	// Flush any remaining events not filling batchSize many blocks
	return flush(stream, buf)
}

func flush(stream pbevents.ExecutionEvents_GetEventsServer, buf *pbevents.GetEventsResponse) error {
	if len(buf.Events) > 0 {
		err := stream.Send(buf)
		if err != nil {
			return err
		}
	}
	return nil
}
