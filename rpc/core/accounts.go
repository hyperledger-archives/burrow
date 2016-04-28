package core

import (
	"fmt"
	acm "github.com/eris-ltd/eris-db/account"
	ctypes "github.com/eris-ltd/eris-db/rpc/core/types"
	. "github.com/tendermint/go-common"
)

func GenPrivAccount() (*ctypes.ResultGenPrivAccount, error) {
	return &ctypes.ResultGenPrivAccount{acm.GenPrivAccount()}, nil
}

// If the account is not known, returns nil, nil.
func GetAccount(address []byte) (*ctypes.ResultGetAccount, error) {
	cache := erisdbApp.GetCheckCache()
	// cache := mempoolReactor.Mempool.GetCache()
	account := cache.GetAccount(address)
	if account == nil {
		return nil, nil
	}
	return &ctypes.ResultGetAccount{account}, nil
}

func GetStorage(address, key []byte) (*ctypes.ResultGetStorage, error) {
	state := erisdbApp.GetState()
	// state := consensusState.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)

	_, value, exists := storageTree.Get(LeftPadWord256(key).Bytes())
	if !exists { // value == nil {
		return &ctypes.ResultGetStorage{key, nil}, nil
	}
	return &ctypes.ResultGetStorage{key, value}, nil
}

func ListAccounts() (*ctypes.ResultListAccounts, error) {
	var blockHeight int
	var accounts []*acm.Account
	state := erisdbApp.GetState()
	blockHeight = state.LastBlockHeight
	state.GetAccounts().Iterate(func(key []byte, value []byte) bool {
		accounts = append(accounts, acm.DecodeAccount(value))
		return false
	})
	return &ctypes.ResultListAccounts{blockHeight, accounts}, nil
}

func DumpStorage(address []byte) (*ctypes.ResultDumpStorage, error) {
	state := erisdbApp.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)
	storageItems := []ctypes.StorageItem{}
	storageTree.Iterate(func(key []byte, value []byte) bool {
		storageItems = append(storageItems, ctypes.StorageItem{key, value})
		return false
	})
	return &ctypes.ResultDumpStorage{storageRoot, storageItems}, nil
}
