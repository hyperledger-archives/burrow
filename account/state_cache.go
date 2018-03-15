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

package account

import (
	"fmt"
	"sort"

	"github.com/hyperledger/burrow/binary"
)

type StateCache struct {
	backend  StateReader
	accounts map[Address]*accountInfo
}

type accountInfo struct {
	account Account
	storage map[binary.Word256]binary.Word256
	removed bool
	updated bool
}

var _ StateWriter = &StateCache{}

func NewStateCache(backend StateReader) *StateCache {
	return &StateCache{
		backend:  backend,
		accounts: make(map[Address]*accountInfo),
	}
}

func (cache *StateCache) GetAccount(address Address) (Account, error) {
	accInfo, err := cache.get(address)
	if err != nil {
		return nil, err
	}
	if accInfo.removed {
		return nil, nil
	}

	if accInfo.account == nil {
		// fill cache
		account, err := cache.backend.GetAccount(address)
		if err != nil {
			return nil, err
		}
		accInfo.account = account
	}
	return accInfo.account, nil
}

func (cache *StateCache) UpdateAccount(account Account) error {
	accInfo, err := cache.get(account.Address())
	if err != nil {
		return err
	}
	if accInfo.removed {
		return fmt.Errorf("UpdateAccount on a removed account %s", account.Address())
	}
	accInfo.account = account
	accInfo.updated = true
	return nil
}

func (cache *StateCache) RemoveAccount(address Address) error {
	accInfo, err := cache.get(address)
	if err != nil {
		return err
	}
	if accInfo.removed {
		fmt.Errorf("RemoveAccount on a removed account %s", address)
	} else {
		accInfo.removed = true
	}
	return nil
}

func (cache *StateCache) GetStorage(address Address, key binary.Word256) (binary.Word256, error) {
	accInfo, err := cache.get(address)
	if err != nil {
		return binary.Zero256, err
	}
	// Check cache
	value, ok := accInfo.storage[key]
	if ok {
		return value, nil
	} else {
		// Load from backend
		value, err := cache.backend.GetStorage(address, key)
		if err != nil {
			return binary.Zero256, err
		}
		accInfo.storage[key] = value
		return value, nil
	}
}

// NOTE: Set value to zero to remove.
func (cache *StateCache) SetStorage(address Address, key binary.Word256, value binary.Word256) error {
	accInfo, err := cache.get(address)
	if err != nil {
		return err
	}
	if accInfo.removed {
		return fmt.Errorf("SetStorage on a removed account %s", address)
	}
	accInfo.storage[key] = value
	accInfo.updated = true
	return nil
}

// Syncs changes to the backend in deterministic order. Sends storage updates before updating
// the account they belong so that storage values can be taken account of in the update.
func (cache *StateCache) Sync(state StateWriter) error {
	var addresses Addresses
	for address := range cache.accounts {
		addresses = append(addresses, address)
	}

	sort.Stable(addresses)
	for _, address := range addresses {
		accInfo := cache.accounts[address]
		if accInfo.removed {
			err := state.RemoveAccount(address)
			if err != nil {
				return err
			}
		} else if accInfo.updated {
			var keys binary.Words256
			for key := range accInfo.storage {
				keys = append(keys, key)
			}
			// First update keys
			sort.Stable(keys)
			for _, key := range keys {
				value := accInfo.storage[key]
				err := state.SetStorage(address, key, value)
				if err != nil {
					return err
				}
			}
			// Update account - this gives backend the opportunity to update StorageRoot after calculating the new
			// value from any storage value updates
			err := state.UpdateAccount(accInfo.account)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Resets the cache to empty initialising the backing map to the same size as the previous iteration.
func (cache *StateCache) Reset(backend StateReader) {
	cache.backend = backend
	cache.accounts = make(map[Address]*accountInfo, len(cache.accounts))
}

func (cache *StateCache) Backend() StateReader {
	return cache.backend
}

// Get the cache accountInfo item creating it if necessary
func (cache *StateCache) get(address Address) (*accountInfo, error) {
	accInfo := cache.accounts[address]
	if accInfo == nil {
		account, err := cache.backend.GetAccount(address)
		if err != nil {
			return nil, err
		}
		accInfo = &accountInfo{
			account: account,
			storage: make(map[binary.Word256]binary.Word256),
		}
		cache.accounts[address] = accInfo
	}
	return accInfo, nil
}
