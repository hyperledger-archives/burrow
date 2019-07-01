package dump

import (
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
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
	st, err := state.MakeGenesisState(testDB(t), &genesis.GenesisDoc{})
	require.NoError(t, err)
	err = Load(mock, st)
	require.NoError(t, err)
	return st
}
