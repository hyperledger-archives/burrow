package rpcevents

import (
	"context"
	"fmt"
	"io"

	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
)

const SubscribeBufferSize = 100

type Provider interface {
	// Get transactions
	IterateStreamEvents(startHeight, endHeight uint64, consumer func(*exec.StreamEvent) error) (err error)
	// Get a particular TxExecution by hash
	TxByHash(txHash []byte) (*exec.TxExecution, error)
}

type executionEventsServer struct {
	eventsProvider Provider
	subscribable   event.Subscribable
	tip            bcm.BlockchainInfo
	logger         *logging.Logger
}

func NewExecutionEventsServer(eventsProvider Provider, subscribable event.Subscribable,
	tip bcm.BlockchainInfo, logger *logging.Logger) ExecutionEventsServer {

	return &executionEventsServer{
		eventsProvider: eventsProvider,
		subscribable:   subscribable,
		tip:            tip,
		logger:         logger.WithScope("NewExecutionEventsServer"),
	}
}

func (ees *executionEventsServer) Tx(ctx context.Context, request *TxRequest) (*exec.TxExecution, error) {
	txe, err := ees.eventsProvider.TxByHash(request.TxHash)
	if err != nil {
		return nil, err
	}
	if txe != nil {
		return txe, nil
	}
	if !request.Wait {
		return nil, fmt.Errorf("transaction with hash %v not found in state", request.TxHash)
	}
	subID := event.GenSubID()
	out, err := ees.subscribable.Subscribe(ctx, subID, exec.QueryForTxExecution(request.TxHash), SubscribeBufferSize)
	if err != nil {
		return nil, err
	}
	defer ees.subscribable.UnsubscribeAll(ctx, subID)
	for msg := range out {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return msg.(*exec.TxExecution), nil
		}
	}
	return nil, fmt.Errorf("subscription waiting for tx %v ended prematurely", request.TxHash)
}

func (ees *executionEventsServer) Stream(request *BlocksRequest, stream ExecutionEvents_StreamServer) error {
	qry, err := query.NewOrEmpty(request.Query)
	if err != nil {
		return fmt.Errorf("could not parse TxExecution query: %v", err)
	}
	return ees.streamBlockEvents(stream.Context(), request.BlockRange, func(ev *exec.StreamEvent) error {
		if qry.Matches(ev.Tagged()) {
			return stream.Send(ev)
		}
		return nil
	})
}

func (ees *executionEventsServer) Events(request *BlocksRequest, stream ExecutionEvents_EventsServer) error {
	qry, err := query.NewOrEmpty(request.Query)
	if err != nil {
		return fmt.Errorf("could not parse Event query: %v", err)
	}
	var response *EventsResponse
	return ees.streamBlockEvents(stream.Context(), request.BlockRange, func(blockEvent *exec.StreamEvent) error {
		switch {
		case blockEvent.BeginBlock != nil:
			response = &EventsResponse{
				Height: blockEvent.BeginBlock.Height,
			}

		case blockEvent.TxExecution != nil:
			// We exclude exceptional transactions - in particular we exclude reverted transactions
			if blockEvent.TxExecution.Exception == nil {
				for _, ev := range blockEvent.TxExecution.Events {
					if qry.Matches(ev.Tagged()) {
						response.Events = append(response.Events, ev)
					}
				}
			}

		case blockEvent.EndBlock != nil && len(response.Events) > 0:
			return stream.Send(response)
		}

		return nil
	})
}

func (ees *executionEventsServer) streamBlockEvents(ctx context.Context, blockRange *BlockRange,
	consumer func(execution *exec.StreamEvent) error) error {

	// Converts the bounds to half-open interval needed
	start, end, streaming := blockRange.Bounds(ees.tip.LastBlockHeight())
	ees.logger.TraceMsg("Streaming blocks", "start", start, "end", end, "streaming", streaming)

	// Pull blocks from state and receive the upper bound (exclusive) on the what we were able to send
	// Set this to start since it will be the start of next streaming batch (if needed)
	start, err := ees.iterateBlockEvents(start, end, consumer)

	// If we are not streaming and all blocks requested were retrieved from state then we are done
	if !streaming && start >= end {
		return err
	}

	return ees.subscribeBlockExecution(ctx, func(block *exec.BlockExecution) error {
		streamEnd := block.Height
		if streamEnd < start {
			// We've managed to receive a block event we already processed directly from state above - wait for next block
			return nil
		}

		finished := !streaming && streamEnd >= end
		if finished {
			// Truncate streamEnd to final end to get exactly the blocks we want from state
			streamEnd = end
		}
		if start < streamEnd {
			// This implies there are some blocks between the previous batchEnd (now start) and the current BlockExecution that
			// we have not emitted so we will pull them from state. This can occur if a block is emitted during/after
			// the initial streaming but before we have subscribed to block events or if we spill BlockExecutions
			// when streaming them and need to catch up
			_, err := ees.iterateBlockEvents(start, streamEnd, consumer)
			if err != nil {
				return err
			}
		}
		if finished {
			return io.EOF
		}
		for _, ev := range block.Events() {
			err = consumer(ev)
			if err != nil {
				return err
			}
		}
		// We've just streamed block so our next start marker is the next block
		start = block.Height + 1
		return nil
	})
}

func (ees *executionEventsServer) subscribeBlockExecution(ctx context.Context, consumer func(*exec.BlockExecution) error) (err error) {
	// Otherwise we need to begin streaming blocks as they are produced
	subID := event.GenSubID()
	// Subscribe to BlockExecution events
	out, err := ees.subscribable.Subscribe(ctx, subID, exec.QueryForBlockExecution(), SubscribeBufferSize)
	if err != nil {
		return err
	}
	defer func() {
		err = ees.subscribable.UnsubscribeAll(context.Background(), subID)
		for range out {
			// flush
		}
	}()

	for msg := range out {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = consumer(msg.(*exec.BlockExecution))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ees *executionEventsServer) iterateBlockEvents(start, end uint64, consumer func(*exec.StreamEvent) error) (lastHeightSeen uint64, err error) {
	err = ees.eventsProvider.IterateStreamEvents(start, end, func(blockEvent *exec.StreamEvent) error {
		if blockEvent.TxExecution == nil {
			// nil safe
			lastHeightSeen = blockEvent.GetEndBlock().GetHeight()
		}
		return consumer(blockEvent)
	})
	// Returns the appropriate starting block for the next stream
	return lastHeightSeen + 1, err
}
