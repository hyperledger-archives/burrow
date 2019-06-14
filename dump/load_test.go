package dump

import (
	"testing"

	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestLoad(t *testing.T) {
	testLoad(t)
}

func BenchmarkLoad(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testLoad(b)
	}
}

func testLoad(tb testing.TB) {
	mock := MockDumpReader{
		accounts: 2000,
		storage:  1000,
		names:    100,
		events:   100000,
	}
	st, err := state.MakeGenesisState(dbm.NewMemDB(), &genesis.GenesisDoc{})
	require.NoError(tb, err)
	err = Load(&mock, st)
	require.NoError(tb, err)
}
