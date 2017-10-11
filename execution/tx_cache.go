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
	"github.com/hyperledger/burrow/word"
)

type TxCache struct {
	backend  acm.StateReader
	accounts map[acm.Address]vmAccountInfo
	storages map[word.Tuple256]word.Word256
}

var _ acm.StateWriter = &TxCache{}

func NewTxCache(backend acm.StateReader) *TxCache {
	return &TxCache{
		backend:  backend,
		accounts: make(map[acm.Address]vmAccountInfo),
		storages: make(map[word.Tuple256]word.Word256),
	}
}

//-------------------------------------
// TxCache.account

func (cache *TxCache) GetAccount(addr acm.Address) (acm.Account, error) {
	acc, removed := cache.accounts[addr].unpack()
	if removed {
		return nil, nil
	} else if acc == nil {
		return cache.backend.GetAccount(addr)
	}
	return acc, nil
}

func (cache *TxCache) UpdateAccount(acc acm.Account) error {
	_, removed := cache.accounts[acc.Address()].unpack()
	if removed {
		return fmt.Errorf("UpdateAccount on a removed account %s", acc.Address())
	}
	cache.accounts[acc.Address()] = vmAccountInfo{acc, false}
	return nil
}

func (cache *TxCache) RemoveAccount(addr acm.Address) error {
	acc, removed := cache.accounts[addr].unpack()
	if removed {
		fmt.Errorf("RemoveAccount on a removed account %s", addr)
	}
	cache.accounts[addr] = vmAccountInfo{acc, true}
	return nil
}

// TxCache.account
//-------------------------------------
// TxCache.storage

func (cache *TxCache) GetStorage(addr acm.Address, key word.Word256) (word.Word256, error) {
	// Check cache
	value, ok := cache.storages[word.Tuple256{First: addr.Word256(), Second: key}]
	if ok {
		return value, nil
	}

	// Load from backend
	return cache.backend.GetStorage(addr, key)
}

// NOTE: Set value to zero to removed from the trie.
func (cache *TxCache) SetStorage(addr acm.Address, key word.Word256, value word.Word256) error {
	_, removed := cache.accounts[addr].unpack()
	if removed {
		fmt.Errorf("SetStorage on a removed account %s", addr)
	}
	cache.storages[word.Tuple256{First: addr.Word256(), Second: key}] = value
	return nil
}

// TxCache.storage
//-------------------------------------

// These updates do not have to be in deterministic order,
// the backend is responsible for ordering updates.
func (cache *TxCache) Sync(backend acm.StateWriter) {
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
	account acm.Account
	removed bool
}

func (accInfo vmAccountInfo) unpack() (acm.Account, bool) {
	return accInfo.account, accInfo.removed
}
