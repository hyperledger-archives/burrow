package core

import (
	"fmt"

	ctypes "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	sm "github.com/eris-ltd/eris-db/manager/eris-mint/state"
	"github.com/eris-ltd/eris-db/txs"
)

func GetName(name string) (*ctypes.ResultGetName, error) {
	st := erisdbApp.GetState()
	entry := st.GetNameRegEntry(name)
	if entry == nil {
		return nil, fmt.Errorf("Name %s not found", name)
	}
	return &ctypes.ResultGetName{entry}, nil
}

func ListNames() (*ctypes.ResultListNames, error) {
	var blockHeight int
	var names []*txs.NameRegEntry
	state := erisdbApp.GetState()
	blockHeight = state.LastBlockHeight
	state.GetNames().Iterate(func(key []byte, value []byte) bool {
		names = append(names, sm.DecodeNameRegEntry(value))
		return false
	})
	return &ctypes.ResultListNames{blockHeight, names}, nil
}
