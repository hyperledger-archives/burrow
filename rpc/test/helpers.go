package test

import (
	"context"
	"testing"

	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/events/pbevents"
	"github.com/hyperledger/burrow/execution/pbtransactor"
	"github.com/hyperledger/burrow/rpc"
	"github.com/stretchr/testify/require"
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

func NewEventsClient(t testing.TB) pbevents.EventsClient {
	conn, err := grpc.Dial(rpc.DefaultGRPCConfig().ListenAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return pbevents.NewEventsClient(conn)
}

func CommittedTxCount(t *testing.T, em event.Emitter) chan int {
	var numTxs int64
	emptyBlocks := 0
	maxEmptyBlocks := 2
	outCh := make(chan int)
	ctx := context.Background()
	ch, err := tendermint.SubscribeNewBlock(ctx, em)
	require.NoError(t, err)

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
		outCh <- int(numTxs)
	}()
	return outCh
}
