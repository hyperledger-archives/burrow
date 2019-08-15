package dump

import (
	"testing"

	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/stretchr/testify/require"
)

func TestDumpLoadCycle(t *testing.T) {
	// Get some initial test data from a mock state
	mock := NewMockSource(100, 1, 20, 10)
	st, err := state.MakeGenesisState(testDB(t), &genesis.GenesisDoc{})
	err = Load(mock, st)
	require.NoError(t, err)

	// We want to check we get the same state after a dump restore, but we cannot compare with the intial loaded state
	// st because mock source does not give dump in same order and IAVL is order-dependent, so we'll chain 2 dump/restores
	// and compare the two resultant states

	// Fresh states to load back into
	stOut1, err := state.MakeGenesisState(testDB(t), &genesis.GenesisDoc{})
	require.NoError(t, err)

	stOut2, err := state.MakeGenesisState(testDB(t), &genesis.GenesisDoc{})
	require.NoError(t, err)

	// First dump from st and load stOut1
	dump := dumpToJSONString(t, st, &bcm.Blockchain{})
	loadDumpFromJSONString(t, stOut1, dump)

	// Now dump from stOut1 and load to stOut2
	dump2 := dumpToJSONString(t, stOut1, &bcm.Blockchain{})
	loadDumpFromJSONString(t, stOut2, dump2)

	require.Equal(t, dump, dump2)
	require.Equal(t, stOut1.Version(), stOut2.Version())
	require.Equal(t, stOut1.Hash(), stOut2.Hash())
}
