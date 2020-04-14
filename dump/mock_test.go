package dump

import (
	"bytes"
	"testing"

	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/require"
)

func TestMockReader(t *testing.T) {
	mock := NewMockSource(100, 5, 20, 1000)
	mock.Mockchain = NewMockchain("TestChain", 0)
	buf := new(bytes.Buffer)
	err := Write(buf, mock, false, All)
	require.NoError(t, err)
	dump := normaliseDump(buf.String())

	st, err := state.MakeGenesisState(testDB(t), &genesis.GenesisDoc{GlobalPermissions: permission.DefaultAccountPermissions})
	require.NoError(t, err)
	loadDumpFromJSONString(t, st, dump)

	dumpOut := normaliseDump(dumpToJSONString(t, st, mock))
	require.Equal(t, dump, dumpOut)
}
