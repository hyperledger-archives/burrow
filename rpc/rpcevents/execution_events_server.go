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
	"github.com/hyperledger/burrow/storage"
)

const SubscribeBufferSize = 100

type Provider interface {
	// Get transactions
	IterateStreamEvents(startHeight, endHeight *uint64, sortOrder storage.SortOrder,
		consumer func(*exec.StreamEvent) error) (err error)
	// Get a particular TxExecution by hash
	TxByHash(txHash []byte) (*exec.TxExecution, error)
}

type executionEventsServer struct {
	eventsProvider Provider
	emitter        *event.Emitter
	tip            bcm.BlockchainInfo
	logger         *logging.Logger
}

func NewExecutionEventsServer(eventsProvider Provider, emitter *event.Emitter,
	tip bcm.BlockchainInfo, logger *logging.Logger) ExecutionEventsServer {

	return &executionEventsServer{
		eventsProvider: eventsProvider,
		emitter:        emitter,
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
	out, err := ees.emitter.Subscribe(ctx, subID, exec.QueryForTxExecution(request.TxHash), SubscribeBufferSize)
	if err != nil {
		return nil, err
	}
	defer ees.emitter.UnsubscribeAll(ctx, subID)
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
	return ees.streamEvents(stream.Context(), request.BlockRange, func(ev *exec.StreamEvent) error {
		if qry.Matches(ev) {
			return stream.Send(ev)
		}
		return nil
	})
}

func (ees *executionEventsServer) Events(request *BlocksRequest, stream ExecutionEvents_EventsServer) error {
	const errHeader = "Events()"
	qry, err := query.NewOrEmpty(request.Query)
	if err != nil {
		return fmt.Errorf("could not parse Event query: %v", err)
	}
	var response *EventsResponse
	var stack exec.TxStack
	return ees.streamEvents(stream.Context(), request.BlockRange, func(sev *exec.StreamEvent) error {
		switch {
		case sev.BeginBlock != nil:
			response = &EventsResponse{
				Height: sev.BeginBlock.Height,
			}

		case sev.EndBlock != nil && len(response.Events) > 0:
			return stream.Send(response)

		default:
			// We need to consume transaction to exclude events belong to an exceptional transaction
			txe, err := stack.Consume(sev)
			if err != nil {
				return fmt.Errorf("%s: %v", errHeader, err)
			}
			if txe != nil && txe.Exception == nil {
				for _, ev := range txe.Events {
					if qry.Matches(ev) {
						response.Events = append(response.Events, ev)
					}
				}
			}
		}

		return nil
	})
}

func (ees *executionEventsServer) streamEvents(ctx context.Context, blockRange *BlockRange,
	consumer func(execution *exec.StreamEvent) error) error {

	start, end, streaming := blockRange.Bounds(ees.tip.LastBlockHeight())
	ees.logger.TraceMsg("Streaming blocks", "start", start, "end", end, "streaming", streaming)

	// Pull blocks from state and receive the upper bound (exclusive) on the what we were able to send
	// Set this to start since it will be the start of next streaming batch (if needed)
	start, err := ees.iterateStreamEvents(start, end, consumer)

	// If we are not streaming and all blocks requested were retrieved from state then we are done
	if !streaming && start > end {
		return err
	}

	return ees.subscribeBlockExecution(ctx, func(block *exec.BlockExecution) error {
		if block.Height < start {
			// We've managed to receive a block event we already processed directly from state above - wait for next block
			return nil
		}
		// Check if we have missed blocks we need to catch up on
		if start < block.Height {
			// We expect start == block.Height when processing consecutive blocks but we may have missed a block by
			// pubsub dropping an event (e.g. under heavy load) - if so we can fill in here. Since we have just
			// received block at block.Height it should be guaranteed that we have stored all blocks <= block.Height
			// in state (we only publish after successful state update).
			catchupEnd := block.Height - 1
			if catchupEnd > end {
				catchupEnd = end
			}
			start, err = ees.iterateStreamEvents(start, catchupEnd, consumer)
			if err != nil {
				return err
			}
		}
		finished := !streaming && block.Height > end
		if finished {
			return io.EOF
		}
		for _, ev := range block.StreamEvents() {
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

func (ees *executionEventsServer) subscribeBlockExecution(ctx context.Context,
	consumer func(*exec.BlockExecution) error) (err error) {
	// Otherwise we need to begin streaming blocks as they are produced
	subID := event.GenSubID()
	// Subscribe to BlockExecution events
	out, err := ees.emitter.Subscribe(ctx, subID, exec.QueryForBlockExecution(), SubscribeBufferSize)
	if err != nil {
		return err
	}
	defer func() {
		err = ees.emitter.UnsubscribeAll(context.Background(), subID)
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

func (ees *executionEventsServer) iterateStreamEvents(startHeight, endHeight uint64,
	consumer func(*exec.StreamEvent) error) (uint64, error) {
	// Assume that we have seen the previous block before start to have ended up here
	// NOTE: this will underflow when start is 0 (as it often will be - and needs to be for restored chains)
	// however we at most underflow by 1 and we always add 1 back on when returning so we get away with this.
	lastHeightSeen := startHeight - 1
	err := ees.eventsProvider.IterateStreamEvents(&startHeight, &endHeight, storage.AscendingSort,
		func(ev *exec.StreamEvent) error {
			if ev.EndBlock != nil {
				lastHeightSeen = ev.EndBlock.GetHeight()
			}
			return consumer(ev)
		})
	// Returns the appropriate _next_ starting block - the one after the one we have seen - from which to stream next
	return lastHeightSeen + 1, err
}
