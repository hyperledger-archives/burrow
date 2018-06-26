package test

import (
	"context"
	"testing"

	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/events/pbevents"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/pbtransactor"
	"github.com/hyperledger/burrow/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmthrgd/go-hex"
	"google.golang.org/grpc"
)

var StrangeLoopBytecode = hex.MustDecodeString(strangeLoopBytecodeHex)

// Recursive call count for UpsieDownsie() function call from strange_loop.sol
// Equals initial call, then depth from 17 -> 34, one for the bounce, then depth from 34 -> 23,
// so... (I didn't say it had to make sense):
const UpsieDownsieCallCount = 1 + (34 - 17) + 1 + (34 - 23)

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

func CreateContract(t testing.TB, cli pbtransactor.TransactorClient,
	inputAccount *pbtransactor.InputAccount) *pbevents.EventDataCall {

	create, err := cli.TransactAndHold(context.Background(), &pbtransactor.TransactParam{
		InputAccount: inputAccount,
		Address:      nil,
		Data:         StrangeLoopBytecode,
		Fee:          2,
		GasLimit:     10000,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(0), create.StackDepth)
	return create
}

func CallContract(t testing.TB, cli pbtransactor.TransactorClient,
	inputAccount *pbtransactor.InputAccount, contractAddress []byte) (call *pbevents.EventDataCall) {

	functionID := abi.FunctionID("UpsieDownsie()")
	call, err := cli.TransactAndHold(context.Background(), &pbtransactor.TransactParam{
		InputAccount: inputAccount,
		Address:      contractAddress,
		Data:         functionID[:],
		Fee:          2,
		GasLimit:     10000,
	})
	require.NoError(t, err)
	return call
}
