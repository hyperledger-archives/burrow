package dump

import (
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	"github.com/hyperledger/burrow/rpc/rpcevents"

	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/storage"

	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	testLoad(t, NewMockSource(100, 10, 20, 1000))
}

func BenchmarkLoad(b *testing.B) {
	for f := 1; f <= 64; f *= 2 {
		b.Run(fmt.Sprintf("factor/%d", f), func(b *testing.B) {
			fmt.Println(f, b.N)
			for n := 0; n < b.N; n++ {
				testLoad(b, NewMockSource(10*f, f, f, 100*f))
			}
		})
	}
}

func testLoad(t testing.TB, mock *MockSource) *state.State {
	st, err := state.MakeGenesisState(testDB(t), &genesis.GenesisDoc{GlobalPermissions: permission.DefaultAccountPermissions})
	require.NoError(t, err)
	err = Load(mock, st)
	require.NoError(t, err)
	return st
}

func TestLoadAndDump(t *testing.T) {
	st, err := state.MakeGenesisState(testDB(t), &genesis.GenesisDoc{GlobalPermissions: permission.DefaultAccountPermissions})
	require.NoError(t, err)
	dir, err := os.Getwd()
	require.NoError(t, err)
	src, err := NewFileReader(path.Join(dir, "test_dump.json"))
	require.NoError(t, err)
	err = Load(src, st)
	require.NoError(t, err)

	// dump and recreate
	for i := 1; i < 10; i++ {
		dumper := NewDumper(st, &bcm.Blockchain{})
		sink := CollectSink{
			Rows: make([]string, 0),
		}
		err = dumper.Transmit(&sink, 0, 0, All)
		require.NoError(t, err)

		st, err = state.MakeGenesisState(testDB(t), &genesis.GenesisDoc{GlobalPermissions: permission.DefaultAccountPermissions, ChainName: fmt.Sprintf("CHAIN #%d", i)})
		require.NoError(t, err)

		err = Load(&sink, st)
		require.NoError(t, err)
	}

	streamEvents := new(exec.StreamEvents)
	eventHeight := uint64(5)
	err = st.IterateStreamEvents(nil, nil, storage.AscendingSort, func(ev *exec.StreamEvent) error {
		streamEvents.StreamEvents = append(streamEvents.StreamEvents, ev)
		if ev.BeginTx != nil {
			require.Equal(t, ev.BeginTx.TxHeader.Origin.Height, eventHeight)
			require.Equal(t, ev.BeginTx.TxHeader.Origin.Index, uint64(2))
			require.Equal(t, ev.BeginTx.TxHeader.Origin.ChainID, "BurrowChain_7DB5BD-5BCE58")
		}
		if ev.Event != nil {
			require.Equal(t, ev.Event.Header.Height, eventHeight)
		}
		return nil
	})
	require.NoError(t, err)

	// Now ensure that the events can be safely consumed by downstream event consumers (e.g. Vent)
	err = rpcevents.ConsumeBlockExecutions(streamEvents, func(be *exec.BlockExecution) error {
		// Events carry their original height in the event header
		require.Equal(t, eventHeight, be.TxExecutions[0].Events[0].Header.Height)
		return nil
	})
	require.Equal(t, io.EOF, err)
}
