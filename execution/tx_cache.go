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

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/word"
)

type TxCache struct {
	backend  acm.GetterAndStorageGetter
	accounts map[acm.Address]vmAccountInfo
	storages map[word.Tuple256]word.Word256
}

var _ evm.State = &TxCache{}

func NewTxCache(backend acm.GetterAndStorageGetter) *TxCache {
	return &TxCache{
		backend:  backend,
		accounts: make(map[acm.Address]vmAccountInfo),
		storages: make(map[word.Tuple256]word.Word256),
	}
}

//-------------------------------------
// TxCache.account

func (cache *TxCache) GetAccount(addr acm.Address) *acm.ConcreteAccount {
	acc, removed := cache.accounts[addr].unpack()
	if removed {
		return nil
	} else if acc == nil {
		acc2 := cache.backend.GetAccount(addr)
		if acc2 != nil {
			return acc2.Copy()
		}
	}
	return acc
}

func (cache *TxCache) UpdateAccount(acc *acm.ConcreteAccount) {
	_, removed := cache.accounts[acc.Address].unpack()
	if removed {
		panic("UpdateAccount on a removed account")
	}
	cache.accounts[acc.Address] = vmAccountInfo{acc, false}
}

func (cache *TxCache) RemoveAccount(addr acm.Address) {
	acc, removed := cache.accounts[addr].unpack()
	if removed {
		panic("RemoveAccount on a removed account")
	}
	cache.accounts[addr] = vmAccountInfo{acc, true}
}

// Creates a 20 byte address and bumps the creator's nonce.
func (cache *TxCache) CreateAccount(creator *acm.ConcreteAccount) *acm.ConcreteAccount {
	// Generate an address
	sequence := creator.Sequence
	creator.Sequence += 1

	addr := txs.NewContractAddress(creator.Address, sequence)

	// Create account from address.
	account, removed := cache.accounts[addr].unpack()
	if removed || account == nil {
		account = &acm.ConcreteAccount{
			Address:     addr,
			Balance:     0,
			Code:        nil,
			Sequence:    0,
			Permissions: cache.GetAccount(permission.GlobalPermissionsAddress).Permissions,
		}
		cache.accounts[addr] = vmAccountInfo{account, false}
		return account
	} else {
		// either we've messed up nonce handling, or sha3 is broken
		panic(fmt.Sprintf("Could not create account, address already exists: %s", addr))
		return nil
	}
}

// TxCache.account
//-------------------------------------
// TxCache.storage

func (cache *TxCache) GetStorage(addr acm.Address, key word.Word256) word.Word256 {
	// Check cache
	value, ok := cache.storages[word.Tuple256{addr.Word256(), key}]
	if ok {
		return value
	}

	// Load from backend
	return cache.backend.GetStorage(addr, key)
}

// NOTE: Set value to zero to removed from the trie.
func (cache *TxCache) SetStorage(addr acm.Address, key word.Word256, value word.Word256) {
	_, removed := cache.accounts[addr].unpack()
	if removed {
		panic("SetStorage() on a removed account")
	}
	cache.storages[word.Tuple256{addr.Word256(), key}] = value
}

// TxCache.storage
//-------------------------------------

// These updates do not have to be in deterministic order,
// the backend is responsible for ordering updates.
func (cache *TxCache) Sync(backend acm.UpdaterAndStorage) {
	// Remove or update storage
	for addrKey, value := range cache.storages {
		addrWord256, key := word.Tuple256Split(addrKey)
		backend.SetStorage(acm.AddressFromWord256(addrWord256), key, value)
	}

	// Remove or update accounts
	for addr, accInfo := range cache.accounts {
		acc, removed := accInfo.unpack()
		if removed {
			backend.RemoveAccount(addr)
		} else {
			backend.UpdateAccount(acc)
		}
	}
}

//-----------------------------------------------------------------------------

type vmAccountInfo struct {
	account *acm.ConcreteAccount
	removed bool
}

func (accInfo vmAccountInfo) unpack() (*acm.ConcreteAccount, bool) {
	return accInfo.account, accInfo.removed
}
