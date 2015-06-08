package client

import (
	"fmt"
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	ms "github.com/mitchellh/mapstructure"
	"github.com/tendermint/tendermint/account"
	"reflect"
)

// Functions used to convert json responses into structs. This is ment to be picked up by
// javascript so type-less return values are good, but when doing go-tests it needs
// to do this work.

func RetEquals(result, concrete, expected interface{}) bool {
	err := ms.Decode(result, concrete)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return reflect.DeepEqual(concrete, expected)
}

func AccountEquals(result interface{}, expected *account.Account) bool {
	return RetEquals(result, &account.Account{}, expected)
}

func AccountsEquals(result interface{}, expected *ep.AccountList) bool {
	return RetEquals(result, &ep.AccountList{}, expected)
}

func StorageEquals(result interface{}, expected *ep.Storage) bool {
	return RetEquals(result, &ep.Storage{}, expected)
}

func StorageAtEquals(result interface{}, expected *ep.StorageItem) bool {
	return RetEquals(result, &ep.StorageItem{}, expected)
}
