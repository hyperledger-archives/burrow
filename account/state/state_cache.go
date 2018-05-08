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

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
)

type Cache interface {
	Writer
	Sync(state Writer) error
	Reset(backend Iterable)
	Flush(state IterableWriter) error
}

type stateCache struct {
	sync.RWMutex
	name     string
	backend  Reader
	accounts map[acm.Address]*accountInfo
}

type accountInfo struct {
	sync.RWMutex
	account acm.Account
	storage map[binary.Word256]binary.Word256
	removed bool
	updated bool
}

type CacheOption func(*stateCache)

// Returns a Cache that wraps an underlying Reader to use on a cache miss, can write to an output Writer
// via Sync. Goroutine safe for concurrent access.
func NewCache(backend Reader, options ...CacheOption) Cache {
	cache := &stateCache{
		backend:  backend,
		accounts: make(map[acm.Address]*accountInfo),
	}
	for _, option := range options {
		option(cache)
	}
	return cache
}

func Name(name string) CacheOption {
	return func(cache *stateCache) {
		cache.name = name
	}
}

func (cache *stateCache) GetAccount(address acm.Address) (acm.Account, error) {
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

func (cache *stateCache) UpdateAccount(account acm.Account) error {
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

func (cache *stateCache) RemoveAccount(address acm.Address) error {
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
func (cache *stateCache) IterateCachedAccount(consumer func(acm.Account) (stop bool)) (stopped bool, err error) {
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

func (cache *stateCache) GetStorage(address acm.Address, key binary.Word256) (binary.Word256, error) {
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
func (cache *stateCache) SetStorage(address acm.Address, key binary.Word256, value binary.Word256) error {
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
func (cache *stateCache) IterateCachedStorage(address acm.Address,
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
func (cache *stateCache) Sync(state Writer) error {
	cache.Lock()
	defer cache.Unlock()
	var addresses acm.Addresses
	for address := range cache.accounts {
		addresses = append(addresses, address)
	}

	sort.Stable(addresses)
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
		accInfo.RUnlock()
	}
	return nil
}

// Resets the cache to empty initialising the backing map to the same size as the previous iteration.
func (cache *stateCache) Reset(backend Iterable) {
	cache.Lock()
	defer cache.Unlock()
	cache.backend = backend
	cache.accounts = make(map[acm.Address]*accountInfo, len(cache.accounts))
}

// Syncs the Cache and Resets it to use as the backend Reader
func (cache *stateCache) Flush(state IterableWriter) error {
	err := cache.Sync(state)
	if err != nil {
		return err
	}
	cache.Reset(state)
	return nil
}

func (cache *stateCache) String() string {
	if cache.name == "" {
		return fmt.Sprintf("StateCache{Length: %v}", len(cache.accounts))
	}
	return fmt.Sprintf("StateCache{Name: %v; Length: %v}", cache.name, len(cache.accounts))
}

// Get the cache accountInfo item creating it if necessary
func (cache *stateCache) get(address acm.Address) (*accountInfo, error) {
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
