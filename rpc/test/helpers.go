package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/events/pbevents"
	"github.com/hyperledger/burrow/execution/pbtransactor"
	"github.com/hyperledger/burrow/rpc"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
)

// Helpers
func NewTransactorClient(t testing.TB) pbtransactor.TransactorClient {
	conn, err := grpc.Dial(rpc.DefaultGRPCConfig().ListenAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return pbtransactor.NewTransactorClient(conn)
}

func NewExecutionEventsClient(t testing.TB) pbevents.ExecutionEventsClient {
	conn, err := grpc.Dial(rpc.DefaultGRPCConfig().ListenAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return pbevents.NewExecutionEventsClient(conn)
}

func CommittedTxCount(t *testing.T, em event.Emitter, committedTxCountIndex *int) chan int {
	var numTxs int64
	emptyBlocks := 0
	maxEmptyBlocks := 2
	outCh := make(chan int)
	ch := make(chan *types.EventDataNewBlock)
	ctx := context.Background()
	subscriber := fmt.Sprintf("committedTxCount_%v", *committedTxCountIndex)
	*committedTxCountIndex++
	require.NoError(t, tendermint.SubscribeNewBlock(ctx, em, subscriber, ch))

	go func() {
		for ed := range ch {
			if ed.Block.NumTxs == 0 {
				emptyBlocks++
			} else {
				emptyBlocks = 0
			}
			if emptyBlocks > maxEmptyBlocks {
				break
			}
			numTxs += ed.Block.NumTxs
			t.Logf("Total TXs committed at block %v: %v (+%v)\n", ed.Block.Height, numTxs, ed.Block.NumTxs)
		}
		require.NoError(t, em.UnsubscribeAll(ctx, subscriber))
		outCh <- int(numTxs)
	}()
	return outCh
}
