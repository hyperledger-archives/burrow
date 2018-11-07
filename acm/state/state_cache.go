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

package state

import (
	"fmt"
	"sort"
	"sync"

	"github.com/hyperledger/burrow/execution/errors"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
)

type Cache struct {
	sync.RWMutex
	name     string
	backend  Reader
	accounts map[crypto.Address]*accountInfo
	readonly bool
}

type accountInfo struct {
	sync.RWMutex
	account acm.Account
	storage map[binary.Word256]binary.Word256
	removed bool
	updated bool
}

type CacheOption func(*Cache) *Cache

// Returns a Cache that wraps an underlying Reader to use on a cache miss, can write to an output Writer
// via Sync. Goroutine safe for concurrent access.
func NewCache(backend Reader, options ...CacheOption) *Cache {
	cache := &Cache{
		backend:  backend,
		accounts: make(map[crypto.Address]*accountInfo),
	}
	for _, option := range options {
		option(cache)
	}
	return cache
}

func Named(name string) CacheOption {
	return func(cache *Cache) *Cache {
		cache.name = name
		return cache
	}
}

var ReadOnly CacheOption = func(cache *Cache) *Cache {
	cache.readonly = true
	return cache
}

func (cache *Cache) GetAccount(address crypto.Address) (acm.Account, error) {
	accInfo, err := cache.get(address)
	if err != nil {
		return nil, err
	}
	accInfo.RLock()
	defer accInfo.RUnlock()
	if accInfo.removed {
		return nil, nil
	}
	return accInfo.account, nil
}

func (cache *Cache) UpdateAccount(account acm.Account) error {
	if cache.readonly {
		return errors.ErrorCodef(errors.ErrorCodeIllegalWrite, "UpdateAccount called on read-only account %v", account.Address())
	}
	accInfo, err := cache.get(account.Address())
	if err != nil {
		return err
	}
	accInfo.Lock()
	defer accInfo.Unlock()
	if accInfo.removed {
		return fmt.Errorf("UpdateAccount on a removed account: %s", account.Address())
	}
	accInfo.account = account
	accInfo.updated = true
	return nil
}

func (cache *Cache) RemoveAccount(address crypto.Address) error {
	if cache.readonly {
		return errors.ErrorCodef(errors.ErrorCodeIllegalWrite, "RemoveAccount called on read-only account %v", address)
	}
	accInfo, err := cache.get(address)
	if err != nil {
		return err
	}
	accInfo.Lock()
	defer accInfo.Unlock()
	if accInfo.removed {
		return fmt.Errorf("RemoveAccount on a removed account: %s", address)
	}
	accInfo.removed = true
	return nil
}

// Iterates over all cached accounts first in cache and then in backend until consumer returns true for 'stop'
func (cache *Cache) IterateCachedAccount(consumer func(acm.Account) (stop bool)) (stopped bool, err error) {
	// Try cache first for early exit
	cache.RLock()
	for _, info := range cache.accounts {
		if consumer(info.account) {
			cache.RUnlock()
			return true, nil
		}
	}
	cache.RUnlock()
	return false, nil
}

func (cache *Cache) GetStorage(address crypto.Address, key binary.Word256) (binary.Word256, error) {
	accInfo, err := cache.get(address)
	if err != nil {
		return binary.Zero256, err
	}
	// Check cache
	accInfo.RLock()
	value, ok := accInfo.storage[key]
	accInfo.RUnlock()
	if !ok {
		accInfo.Lock()
		defer accInfo.Unlock()
		value, ok = accInfo.storage[key]
		if !ok {
			// Load from backend
			value, err = cache.backend.GetStorage(address, key)
			if err != nil {
				return binary.Zero256, err
			}
			accInfo.storage[key] = value
		}
	}
	return value, nil
}

// NOTE: Set value to zero to remove.
func (cache *Cache) SetStorage(address crypto.Address, key binary.Word256, value binary.Word256) error {
	if cache.readonly {
		return errors.ErrorCodef(errors.ErrorCodeIllegalWrite, "SetStorage called on read-only account %v", address)
	}
	accInfo, err := cache.get(address)
	accInfo.Lock()
	defer accInfo.Unlock()
	if err != nil {
		return err
	}
	if accInfo.removed {
		return fmt.Errorf("SetStorage on a removed account: %s", address)
	}
	accInfo.storage[key] = value
	accInfo.updated = true
	return nil
}

// Iterates over all cached storage items first in cache and then in backend until consumer returns true for 'stop'
func (cache *Cache) IterateCachedStorage(address crypto.Address,
	consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {
	accInfo, err := cache.get(address)
	if err != nil {
		return false, err
	}
	accInfo.RLock()
	// Try cache first for early exit
	for key, value := range accInfo.storage {
		if consumer(key, value) {
			accInfo.RUnlock()
			return true, nil
		}
	}
	accInfo.RUnlock()
	return false, nil
}

// Syncs changes to the backend in deterministic order. Sends storage updates before updating
// the account they belong so that storage values can be taken account of in the update.
func (cache *Cache) Sync(state Writer) error {
	if cache.readonly {
		return nil
	}
	cache.Lock()
	defer cache.Unlock()
	var addresses crypto.Addresses
	for address := range cache.accounts {
		addresses = append(addresses, address)
	}

	sort.Sort(addresses)
	for _, address := range addresses {
		accInfo := cache.accounts[address]
		accInfo.RLock()
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
			sort.Sort(keys)
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
		accInfo.RUnlock()
	}
	return nil
}

// Resets the cache to empty initialising the backing map to the same size as the previous iteration.
func (cache *Cache) Reset(backend Reader) {
	cache.Lock()
	defer cache.Unlock()
	cache.backend = backend
	cache.accounts = make(map[crypto.Address]*accountInfo, len(cache.accounts))
}

// Syncs the Cache to output and Resets it to use backend as Reader
func (cache *Cache) Flush(output Writer, backend Reader) error {
	err := cache.Sync(output)
	if err != nil {
		return err
	}
	cache.Reset(backend)
	return nil
}

func (cache *Cache) String() string {
	if cache.name == "" {
		return fmt.Sprintf("StateCache{Length: %v}", len(cache.accounts))
	}
	return fmt.Sprintf("StateCache{Name: %v; Length: %v}", cache.name, len(cache.accounts))
}

// Get the cache accountInfo item creating it if necessary
func (cache *Cache) get(address crypto.Address) (*accountInfo, error) {
	cache.RLock()
	accInfo := cache.accounts[address]
	cache.RUnlock()
	if accInfo == nil {
		cache.Lock()
		defer cache.Unlock()
		accInfo = cache.accounts[address]
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
	}
	return accInfo, nil
}
