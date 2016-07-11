// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

// Accounts is part of the pipe for ErisMint and provides the implementation
// for the pipe to call into the ErisMint application
package erismint

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	tendermint_common "github.com/tendermint/go-common"

	account "github.com/eris-ltd/eris-db/account"
	core_types "github.com/eris-ltd/eris-db/core/types"
	definitions "github.com/eris-ltd/eris-db/definitions"
	event "github.com/eris-ltd/eris-db/event"
)

// NOTE [ben] Compiler check to ensure Accounts successfully implements
// eris-db/definitions.Accounts
var _ definitions.Accounts = (*accounts)(nil)

// The accounts struct has methods for working with accounts.
type accounts struct {
	erisMint      *ErisMint
	filterFactory *event.FilterFactory
}

func newAccounts(erisMint *ErisMint) *accounts {
	ff := event.NewFilterFactory()

	ff.RegisterFilterPool("code", &sync.Pool{
		New: func() interface{} {
			return &AccountCodeFilter{}
		},
	})

	ff.RegisterFilterPool("balance", &sync.Pool{
		New: func() interface{} {
			return &AccountBalanceFilter{}
		},
	})

	return &accounts{erisMint, ff}

}

// Generate a new Private Key Account.
func (this *accounts) GenPrivAccount() (*account.PrivAccount, error) {
	pa := account.GenPrivAccount()
	return pa, nil
}

// Generate a new Private Key Account.
func (this *accounts) GenPrivAccountFromKey(privKey []byte) (
	*account.PrivAccount, error) {
	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not 64 bytes long.")
	}
	fmt.Printf("PK BYTES FROM ACCOUNTS: %x\n", privKey)
	pa := account.GenPrivAccountFromPrivKeyBytes(privKey)
	return pa, nil
}

// Get all accounts.
func (this *accounts) Accounts(fda []*event.FilterData) (
	*core_types.AccountList, error) {
	accounts := make([]*account.Account, 0)
	state := this.erisMint.GetState()
	filter, err := this.filterFactory.NewFilter(fda)
	if err != nil {
		return nil, fmt.Errorf("Error in query: " + err.Error())
	}
	state.GetAccounts().Iterate(func(key, value []byte) bool {
		acc := account.DecodeAccount(value)
		if filter.Match(acc) {
			accounts = append(accounts, acc)
		}
		return false
	})
	return &core_types.AccountList{accounts}, nil
}

// Get an account.
func (this *accounts) Account(address []byte) (*account.Account, error) {
	cache := this.erisMint.GetState() // NOTE: we want to read from mempool!
	acc := cache.GetAccount(address)
	if acc == nil {
		acc = this.newAcc(address)
	}
	return acc, nil
}

// Get the value stored at 'key' in the account with address 'address'
// Both the key and value is returned.
func (this *accounts) StorageAt(address, key []byte) (*core_types.StorageItem,
	error) {
	state := this.erisMint.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return &core_types.StorageItem{key, []byte{}}, nil
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)

	_, value, _ := storageTree.Get(tendermint_common.LeftPadWord256(key).Bytes())
	if value == nil {
		return &core_types.StorageItem{key, []byte{}}, nil
	}
	return &core_types.StorageItem{key, value}, nil
}

// Get the storage of the account with address 'address'.
func (this *accounts) Storage(address []byte) (*core_types.Storage, error) {

	state := this.erisMint.GetState()
	account := state.GetAccount(address)
	storageItems := make([]core_types.StorageItem, 0)
	if account == nil {
		return &core_types.Storage{nil, storageItems}, nil
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)

	storageTree.Iterate(func(key, value []byte) bool {
		storageItems = append(storageItems, core_types.StorageItem{
			key, value})
		return false
	})
	return &core_types.Storage{storageRoot, storageItems}, nil
}

// Create a new account.
func (this *accounts) newAcc(address []byte) *account.Account {
	return &account.Account{
		Address:     address,
		PubKey:      nil,
		Sequence:    0,
		Balance:     0,
		Code:        nil,
		StorageRoot: nil,
	}
}

// Filter for account code.
// Ops: == or !=
// Could be used to match against nil, to see if an account is a contract account.
type AccountCodeFilter struct {
	op    string
	value []byte
	match func([]byte, []byte) bool
}

func (this *AccountCodeFilter) Configure(fd *event.FilterData) error {
	op := fd.Op
	val, err := hex.DecodeString(fd.Value)

	if err != nil {
		return fmt.Errorf("Wrong value type.")
	}
	if op == "==" {
		this.match = func(a, b []byte) bool {
			return bytes.Equal(a, b)
		}
	} else if op == "!=" {
		this.match = func(a, b []byte) bool {
			return !bytes.Equal(a, b)
		}
	} else {
		return fmt.Errorf("Op: " + this.op + " is not supported for 'code' filtering")
	}
	this.op = op
	this.value = val
	return nil
}

func (this *AccountCodeFilter) Match(v interface{}) bool {
	acc, ok := v.(*account.Account)
	if !ok {
		return false
	}
	return this.match(acc.Code, this.value)
}

// Filter for account balance.
// Ops: All
type AccountBalanceFilter struct {
	op    string
	value int64
	match func(int64, int64) bool
}

func (this *AccountBalanceFilter) Configure(fd *event.FilterData) error {
	val, err := event.ParseNumberValue(fd.Value)
	if err != nil {
		return err
	}
	match, err2 := event.GetRangeFilter(fd.Op, "balance")
	if err2 != nil {
		return err2
	}
	this.match = match
	this.op = fd.Op
	this.value = val
	return nil
}

func (this *AccountBalanceFilter) Match(v interface{}) bool {
	acc, ok := v.(*account.Account)
	if !ok {
		return false
	}
	return this.match(int64(acc.Balance), this.value)
}
