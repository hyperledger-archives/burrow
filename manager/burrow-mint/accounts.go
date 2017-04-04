// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package burrowmint

// Accounts is part of the pipe for BurrowMint and provides the implementation
// for the pipe to call into the BurrowMint application

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	account "github.com/monax/burrow/account"
	core_types "github.com/monax/burrow/core/types"
	definitions "github.com/monax/burrow/definitions"
	event "github.com/monax/burrow/event"
	word256 "github.com/monax/burrow/word256"
)

// NOTE [ben] Compiler check to ensure Accounts successfully implements
// burrow/definitions.Accounts
var _ definitions.Accounts = (*accounts)(nil)

// The accounts struct has methods for working with accounts.
type accounts struct {
	burrowMint      *BurrowMint
	filterFactory *event.FilterFactory
}

func newAccounts(burrowMint *BurrowMint) *accounts {
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

	return &accounts{burrowMint, ff}

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
	state := this.burrowMint.GetState()
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
	cache := this.burrowMint.GetState() // NOTE: we want to read from mempool!
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
	state := this.burrowMint.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return &core_types.StorageItem{key, []byte{}}, nil
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)

	_, value, _ := storageTree.Get(word256.LeftPadWord256(key).Bytes())
	if value == nil {
		return &core_types.StorageItem{key, []byte{}}, nil
	}
	return &core_types.StorageItem{key, value}, nil
}

// Get the storage of the account with address 'address'.
func (this *accounts) Storage(address []byte) (*core_types.Storage, error) {

	state := this.burrowMint.GetState()
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

// Function for matching accounts against filter data.
func (this *accounts) matchBlock(block, fda []*event.FilterData) bool {
	return false
}
