package forensics

import (
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

// This serves as a testbed for looking at non-deterministic burrow instances capture from the wild
// Put the path to 'good' and 'bad' burrow directories here (containing the config files and .burrow dir)

func TestStateComp(t *testing.T) {
	st1 := state.NewState(dbm.NewMemDB())
	_, _, err := st1.Update(func(ws state.Updatable) error {
		return ws.UpdateAccount(acm.NewAccountFromSecret("1"))
	})
	require.NoError(t, err)
	_, _, err = st1.Update(func(ws state.Updatable) error {
		return ws.UpdateAccount(acm.NewAccountFromSecret("2"))
	})
	require.NoError(t, err)

	db2 := dbm.NewMemDB()
	st2, err := st1.Copy(db2)
	require.NoError(t, err)
	_, _, err = st2.Update(func(ws state.Updatable) error {
		return ws.UpdateAccount(acm.NewAccountFromSecret("3"))
	})
	require.NoError(t, err)

	err = CompareState(st2, st1, 1)
	require.Error(t, err)
}
