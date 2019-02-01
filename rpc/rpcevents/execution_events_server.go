package rpcevents

import (
	"context"
	"fmt"

	"github.com/tendermint/tendermint/types"

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
	GetTxsAtHeight(height uint64) ([]*exec.TxExecution, error)
	// Get a particular TxExecution by hash
	GetTxByHash(txHash []byte) (*exec.TxExecution, error)
	// Get events between startKey (inclusive) and endKey (exclusive) - i.e. the half open interval [start, end)
	GetTxs(startHeight, endHeight uint64, consumer func(*exec.TxExecution) error) (err error)
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
	txExecutions, err := ees.eventsProvider.GetTxsAtHeight(request.GetHeight())
	if err != nil {
		return nil, err
	}
	// If block of TxExecutions found at height then this is a previous block so make and return
	if txExecutions != nil {
		return ees.makeBlock(request.Height, txExecutions)
	}
	// Otherwise see if we should wait
	if !request.Wait {
		if ees.tip.LastBlockHeight() < request.Height {
			return nil, fmt.Errorf("block at height %v not yet produced (last block height: %v)",
				request.Height, ees.tip.LastBlockHeight())
		}
		return nil, fmt.Errorf("block at height %v not found in state but should have been! (last block height: %v)",
			request.Height, ees.tip.LastBlockHeight())
	}
	var block *exec.BlockExecution
	err = ees.streamTxsByBlock(ctx, &BlockRange{End: StreamBound()},
		func(height uint64, txExecutions []*exec.TxExecution) error {
			if height == request.Height {
				block, err = ees.makeBlock(height, txExecutions)
				if err != nil {
					return err
				}
				return io.EOF
			}
			return nil
		})
	if err != io.EOF {
		return nil, err
	}
	return block, nil
}

func (ees *executionEventsServer) GetBlocks(request *BlocksRequest, stream ExecutionEvents_GetBlocksServer) error {
	qry, err := query.NewOrEmpty(request.Query)
	if err != nil {
		return fmt.Errorf("could not parse BlockExecution query: %v", err)
	}
	return ees.streamTxsByBlock(stream.Context(), request.BlockRange,
		func(height uint64, txExecutions []*exec.TxExecution) error {
			block, err := ees.makeBlock(height, txExecutions)
			if err != nil {
				return err
			}
			if qry.Matches(block.Tagged()) {
				return flush(stream, block)
			}
			return nil
		})
}

func (ees *executionEventsServer) GetTx(ctx context.Context, request *GetTxRequest) (*exec.TxExecution, error) {
	txe, err := ees.eventsProvider.GetTxByHash(request.TxHash)
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
	qry, err := query.NewOrEmpty(request.Query)
	if err != nil {
		return fmt.Errorf("could not parse TxExecution query: %v", err)
	}
	return ees.streamTxsByBlock(stream.Context(), request.BlockRange,
		func(height uint64, txExecutions []*exec.TxExecution) error {
			txs := filterTxs(txExecutions, qry)
			if len(txs) > 0 {
				response := &GetTxsResponse{
					Height:       height,
					TxExecutions: txs,
				}
				return flush(stream, response)
			}
			return nil
		})
}

func (ees *executionEventsServer) GetEvents(request *BlocksRequest, stream ExecutionEvents_GetEventsServer) error {
	qry, err := query.NewOrEmpty(request.Query)
	if err != nil {
		return fmt.Errorf("could not parse Event query: %v", err)
	}
	return ees.streamTxsByBlock(stream.Context(), request.BlockRange, func(height uint64, txe []*exec.TxExecution) error {
		evs := filterEvents(txe, qry)
		if len(evs) == 0 {
			return nil
		}
		response := &GetEventsResponse{
			Height: height,
			Events: evs,
		}
		return flush(stream, response)
	})
}
func (ees *executionEventsServer) streamTxsByBlock(ctx context.Context, blockRange *BlockRange,
	consumer func(height uint64, txes []*exec.TxExecution) error) error {
	var txExecutions []*exec.TxExecution
	var height uint64
	return ees.streamTxs(ctx, blockRange, func(txe *exec.TxExecution) error {
		if txe.Height > height {
			if len(txExecutions) > 0 {
				err := consumer(height, txExecutions)
				if err != nil {
					return err
				}
			}
			txExecutions = txExecutions[:0]
			height = txe.Height
		} else {
			txExecutions = append(txExecutions, txe)
		}
		return nil
	})
}

func (ees *executionEventsServer) streamTxs(ctx context.Context, blockRange *BlockRange,
	consumer func(*exec.TxExecution) error) error {

	// Converts the bounds to half-open interval needed
	start, end, streaming := blockRange.Bounds(ees.tip.LastBlockHeight())
	ees.logger.TraceMsg("Streaming blocks", "start", start, "end", end, "streaming", streaming)

	// Pull blocks from state and receive the upper bound (exclusive) on the what we were able to send
	// Set this to start since it will be the start of next streaming batch (if needed)
	start, err := ees.iterateTxs(start, end, consumer)

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
				_, err := ees.iterateTxs(start, streamEnd, consumer)
				if err != nil {
					return err
				}
			}
			if finished {
				return nil
			}
			for _, txe := range block.TxExecutions {
				err = consumer(txe)
				if err != nil {
					return err
				}
			}
			// We've just streamed block so our next start marker is the next block
			start = block.Height + 1
		}
	}

	return nil
}

// Converts blocks into responses and streams them returning the height one greater than the last seen block
// that can be used as next start point (half-open interval)
func (ees *executionEventsServer) iterateTxs(start, end uint64, consumer func(*exec.TxExecution) error) (uint64, error) {
	var streamErr error
	var lastHeightSeen uint64

	err := ees.eventsProvider.GetTxs(start, end,
		func(txe *exec.TxExecution) error {
			lastHeightSeen = txe.Height
			return consumer(txe)
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

func (ees *executionEventsServer) makeBlock(height uint64, txExecutions []*exec.TxExecution) (*exec.BlockExecution, error) {
	header, err := ees.tip.GetBlockHeader(height)
	if err != nil {
		return nil, err
	}

	abciHeader := types.TM2PB.Header(header)
	return &exec.BlockExecution{
		BlockHeader:  &abciHeader,
		Height:       height,
		TxExecutions: txExecutions,
	}, nil
}

func filterTxs(txExecutions []*exec.TxExecution, qry query.Query) []*exec.TxExecution {
	var txs []*exec.TxExecution
	for _, txe := range txExecutions {
		if qry.Matches(txe.Tagged()) {
			txs = append(txs, txe)
		}
	}
	return txs
}

func filterEvents(txExecutions []*exec.TxExecution, qry query.Query) []*exec.Event {
	var evs []*exec.Event
	for _, txe := range txExecutions {
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
