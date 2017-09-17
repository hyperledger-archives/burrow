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

package execution

// Accounts is part of the pipe for BurrowMint and provides the implementation
// for the pipe to call into the BurrowMint application

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/event"
	word256 "github.com/hyperledger/burrow/word"
)

// The accounts struct has methods for working with accounts.
type accounts struct {
	state         *State
	filterFactory *event.FilterFactory
}

// Accounts
type AccountList struct {
	Accounts []*account.ConcreteAccount `json:"accounts"`
}

// A contract account storage item.
type StorageItem struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

// Account storage
type Storage struct {
	StorageRoot  []byte        `json:"storage_root"`
	StorageItems []StorageItem `json:"storage_items"`
}

// TODO: [Silas] there are various notes about using mempool (which I guess translates to CheckTx cache). We need
// to understand if this is the right thing to do, since we cannot guarantee stability of the check cache it doesn't
// seem like the right thing to do....
func newAccounts(state *State) *accounts {
	filterFactory := event.NewFilterFactory()

	filterFactory.RegisterFilterPool("code", &sync.Pool{
		New: func() interface{} {
			return &AccountCodeFilter{}
		},
	})

	filterFactory.RegisterFilterPool("balance", &sync.Pool{
		New: func() interface{} {
			return &AccountBalanceFilter{}
		},
	})

	return &accounts{
		state:         state,
		filterFactory: filterFactory,
	}
}

// Generate a new Private Key Account.
func (accs *accounts) GenPrivAccount() (*account.ConcretePrivateAccount, error) {
	pa := account.GenPrivAccount().Unwrap()
	return pa, nil
}

// Generate a new Private Key Account.
func (accs *accounts) GenPrivAccountFromKey(privKey []byte) (
	*account.ConcretePrivateAccount, error) {
	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not 64 bytes long.")
	}
	fmt.Printf("PK BYTES FROM ACCOUNTS: %x\n", privKey)
	pa := account.GenPrivAccountFromPrivKeyBytes(privKey).Unwrap()
	return pa, nil
}

// Get all accounts.
func (accs *accounts) Accounts(fda []*event.FilterData) (
	*AccountList, error) {
	accounts := make([]*account.ConcreteAccount, 0)
	filter, err := accs.filterFactory.NewFilter(fda)
	if err != nil {
		return nil, fmt.Errorf("Error in query: " + err.Error())
	}
	accs.state.GetAccounts().Iterate(func(key, value []byte) bool {
		acc := account.DecodeAccount(value)
		if filter.Match(acc) {
			accounts = append(accounts, acc)
		}
		return false
	})
	return &AccountList{accounts}, nil
}

// Get an account.
func (accs *accounts) Account(address account.Address) (*account.ConcreteAccount, error) {
	acc := accs.state.GetAccount(address) // NOTE: we want to read from mempool!
	if acc == nil {
		acc = accs.newAcc(address)
	}
	return acc, nil
}

// Get the value stored at 'key' in the account with address 'address'
// Both the key and value is returned.
func (accs *accounts) StorageAt(address account.Address, key []byte) (*StorageItem,
	error) {
	acc := accs.state.GetAccount(address)
	if acc == nil {
		return &StorageItem{key, []byte{}}, nil
	}
	storageRoot := acc.StorageRoot
	storageTree := accs.state.LoadStorage(storageRoot)

	_, value, _ := storageTree.Get(word256.LeftPadWord256(key).Bytes())
	if value == nil {
		return &StorageItem{key, []byte{}}, nil
	}
	return &StorageItem{key, value}, nil
}

// Get the storage of the account with address 'address'.
func (accs *accounts) Storage(address account.Address) (*Storage, error) {
	state := accs.state
	acc := state.GetAccount(address)
	storageItems := make([]StorageItem, 0)
	if acc == nil {
		return &Storage{nil, storageItems}, nil
	}
	storageRoot := acc.StorageRoot
	storageTree := state.LoadStorage(storageRoot)

	storageTree.Iterate(func(key, value []byte) bool {
		storageItems = append(storageItems, StorageItem{
			key, value})
		return false
	})
	return &Storage{storageRoot, storageItems}, nil
}

// Create a new account.
func (accs *accounts) newAcc(address account.Address) *account.ConcreteAccount {
	return &account.ConcreteAccount{
		Address:     address,
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
	acc, ok := v.(*account.ConcreteAccount)
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
	acc, ok := v.(*account.ConcreteAccount)
	if !ok {
		return false
	}
	return this.match(int64(acc.Balance), this.value)
}
