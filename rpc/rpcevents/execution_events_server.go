package rpcevents

import (
	"context"
	"fmt"

	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"google.golang.org/grpc"
)

const SubscribeBufferSize = 100

type Provider interface {
	// Get a particular BlockExecution
	GetBlock(height uint64) (*exec.BlockExecution, error)
	// Get a partiualr TxExecution by hash
	GetTx(txHash []byte) (*exec.TxExecution, error)
	// Get events between startKey (inclusive) and endKey (exclusive) - i.e. the half open interval [start, end)
	GetBlocks(startHeight, endHeight uint64, consumer func(*exec.BlockExecution) (stop bool)) (stopped bool, err error)
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

func (ees *executionEventsServer) GetBlock(ctx context.Context, request *GetBlockRequest) (*exec.BlockExecution, error) {
	be, err := ees.eventsProvider.GetBlock(request.GetHeight())
	if err != nil {
		return nil, err
	}
	if be != nil {
		return be, nil
	}
	if !request.Wait {
		if ees.tip.LastBlockHeight() < request.Height {
			return nil, fmt.Errorf("block at height %v not yet produced (last block height: %v)",
				request.Height, ees.tip.LastBlockHeight())
		}
		return nil, fmt.Errorf("block at height %v not found in state but should have been! (last block height: %v)",
			request.Height, ees.tip.LastBlockHeight())
	}
	err = ees.streamBlocks(ctx, &BlockRange{End: StreamBound()}, func(block *exec.BlockExecution) error {
		if block.Height == request.Height {
			be = block
			return io.EOF
		}
		return nil
	})
	if err != io.EOF {
		return nil, err
	}
	return be, nil
}

func (ees *executionEventsServer) GetBlocks(request *BlocksRequest, stream ExecutionEvents_GetBlocksServer) error {
	qry, err := query.NewBuilder(request.Query).Query()
	if err != nil {
		return fmt.Errorf("could not parse BlockExecution query: %v", err)
	}
	return ees.streamBlocks(stream.Context(), request.BlockRange, func(block *exec.BlockExecution) error {
		if qry.Matches(block.Tagged()) {
			return flush(stream, block)
		}
		return nil
	})
}

func (ees *executionEventsServer) GetTx(ctx context.Context, request *GetTxRequest) (*exec.TxExecution, error) {
	txe, err := ees.eventsProvider.GetTx(request.TxHash)
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

func (ees *executionEventsServer) GetTxs(request *BlocksRequest, stream ExecutionEvents_GetTxsServer) error {
	qry, err := query.NewBuilder(request.Query).Query()
	if err != nil {
		return fmt.Errorf("could not parse TxExecution query: %v", err)
	}
	return ees.streamBlocks(stream.Context(), request.BlockRange, func(block *exec.BlockExecution) error {
		txs := filterTxs(block, qry)
		if len(txs) > 0 {
			response := &GetTxsResponse{
				Height:       block.Height,
				TxExecutions: txs,
			}
			return flush(stream, response)
		}
		return nil
	})
}

func (ees *executionEventsServer) GetEvents(request *BlocksRequest, stream ExecutionEvents_GetEventsServer) error {
	qry, err := query.NewBuilder(request.Query).Query()
	if err != nil {
		return fmt.Errorf("could not parse Event query: %v", err)
	}
	return ees.streamBlocks(stream.Context(), request.BlockRange, func(block *exec.BlockExecution) error {
		evs := filterEvents(block, qry)
		if len(evs) == 0 {
			return nil
		}
		response := &GetEventsResponse{
			Height: block.Height,
			Events: evs,
		}
		return flush(stream, response)
	})
}

func (ees *executionEventsServer) streamBlocks(ctx context.Context, blockRange *BlockRange,
	consumer func(*exec.BlockExecution) error) error {

	// Converts the bounds to half-open interval needed
	start, end, streaming := blockRange.Bounds(ees.tip.LastBlockHeight())
	ees.logger.TraceMsg("Streaming blocks", "start", start, "end", end, "streaming", streaming)

	// Pull blocks from state and receive the upper bound (exclusive) on the what we were able to send
	// Set this to start since it will be the start of next streaming batch (if needed)
	start, err := ees.iterateBlocks(start, end, consumer)

	// If we are not streaming and all blocks requested were retrieved from state then we are done
	if !streaming && start == end {
		return err
	}

	// Otherwise we need to begin streaming blocks as they are produced
	subID := event.GenSubID()
	// Subscribe to BlockExecution events
	out, err := ees.subscribable.Subscribe(ctx, subID, exec.QueryForBlockExecutionFromHeight(end),
		SubscribeBufferSize)
	if err != nil {
		return err
	}
	defer ees.subscribable.UnsubscribeAll(context.Background(), subID)

	for msg := range out {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			block := msg.(*exec.BlockExecution)
			streamEnd := block.Height

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
				_, err := ees.iterateBlocks(start, streamEnd, consumer)
				if err != nil {
					return err
				}
			}
			if finished {
				return nil
			}
			err = consumer(block)
			if err != nil {
				return err
			}
			// We've just streamed block so our next start marker is the next block
			start = block.Height + 1
		}
	}

	return nil
}

// Converts blocks into responses and streams them returning the height one greater than the last seen block
// that can be used as next start point (half-open interval)
func (ees *executionEventsServer) iterateBlocks(start, end uint64, consumer func(*exec.BlockExecution) error) (uint64, error) {
	var streamErr error
	var lastHeightSeen uint64

	_, err := ees.eventsProvider.GetBlocks(start, end,
		func(be *exec.BlockExecution) (stop bool) {
			lastHeightSeen = be.Height
			streamErr = consumer(be)
			if streamErr != nil {
				return true
			}
			return false
		})

	if err != nil {
		return 0, err
	}
	if streamErr != nil {
		return 0, streamErr
	}
	// Returns the appropriate starting block for the next stream
	return lastHeightSeen + 1, nil
}

func filterTxs(be *exec.BlockExecution, qry query.Query) []*exec.TxExecution {
	var txs []*exec.TxExecution
	for _, txe := range be.TxExecutions {
		if qry.Matches(txe.Tagged()) {
			txs = append(txs, txe)
		}
	}
	return txs
}

func filterEvents(be *exec.BlockExecution, qry query.Query) []*exec.Event {
	var evs []*exec.Event
	for _, txe := range be.TxExecutions {
		if txe.Exception == nil {
			for _, ev := range txe.Events {
				if qry.Matches(ev.Tagged()) {
					evs = append(evs, ev)
				}
			}
		}
	}
	return evs
}

func flush(stream grpc.Stream, buf proto.Message) error {
	if buf != nil {
		err := stream.SendMsg(buf)
		if err != nil {
			return err
		}
	}
	return nil
}
