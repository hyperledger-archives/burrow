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
	"bytes"
	"fmt"
	"sort"

	acm "github.com/hyperledger/burrow/account"
	. "github.com/hyperledger/burrow/binary"

	"github.com/tendermint/merkleeyes/iavl"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/merkle"
	"sync"
)

func makeStorage(db dbm.DB, root []byte) merkle.Tree {
	storage := iavl.NewIAVLTree(1024, db)
	storage.Load(root)
	return storage
}

var _ acm.StateWriter = &BlockCache{}

var _ acm.StateIterable = &BlockCache{}

// TODO: BlockCache badly needs a rewrite to remove database sharing with State and make it communicate using the
// Account interfaces like a proper person. As well as other oddities of decoupled storage and account state

// The blockcache helps prevent unnecessary IAVLTree updates and garbage generation.
type BlockCache struct {
	// We currently provide the RPC layer with access to read-only access to BlockCache via the StateIterable interface
	// on BatchExecutor. However since read-only operations generate writes to the BlockCache in the current design
	// we need a mutex here. Otherwise BlockCache ought to be used within a component that is responsible for serialising
	// the operations on the BlockCache.
	sync.RWMutex
	db       dbm.DB
	backend  *State
	accounts map[acm.Address]accountInfo
	storages map[acm.Address]map[Word256]storageInfo
	names    map[string]nameInfo
}

func NewBlockCache(backend *State) *BlockCache {
	return &BlockCache{
		// TODO: This is bad and probably the cause of various panics. Accounts themselves are written
		// to the State 'backend' but updates to storage just skip that and write directly to the database
		db:       backend.db,
		backend:  backend,
		accounts: make(map[acm.Address]accountInfo),
		storages: make(map[acm.Address]map[Word256]storageInfo),
		names:    make(map[string]nameInfo),
	}
}

func (cache *BlockCache) State() *State {
	return cache.backend
}

//-------------------------------------
// BlockCache.account

func (cache *BlockCache) GetAccount(addr acm.Address) (acm.Account, error) {
	acc, _, removed, _ := cache.accounts[addr].unpack()
	if removed {
		return nil, nil
	} else if acc != nil {
		return acc, nil
	} else {
		acc, err := cache.backend.GetAccount(addr)
		if err != nil {
			return nil, err
		}
		cache.Lock()
		defer cache.Unlock()
		cache.accounts[addr] = accountInfo{acc, nil, false, false}
		return acc, nil
	}
}

func (cache *BlockCache) UpdateAccount(acc acm.Account) error {
	cache.Lock()
	defer cache.Unlock()
	addr := acc.Address()
	_, storage, removed, _ := cache.accounts[addr].unpack()
	if removed {
		return fmt.Errorf("UpdateAccount on a removed account %s", addr)
	}
	cache.accounts[addr] = accountInfo{acc, storage, false, true}
	return nil
}

func (cache *BlockCache) RemoveAccount(addr acm.Address) error {
	cache.Lock()
	defer cache.Unlock()
	_, _, removed, _ := cache.accounts[addr].unpack()
	if removed {
		return fmt.Errorf("RemoveAccount on a removed account %s", addr)
	}
	cache.accounts[addr] = accountInfo{nil, nil, true, false}
	return nil
}

func (cache *BlockCache) IterateAccounts(consumer func(acm.Account) (stop bool)) (bool, error) {
	cache.RLock()
	defer cache.RUnlock()
	for _, info := range cache.accounts {
		if consumer(info.account) {
			return true, nil
		}
	}
	return cache.backend.IterateAccounts(consumer)
}

// BlockCache.account
//-------------------------------------
// BlockCache.storage

func (cache *BlockCache) GetStorage(addr acm.Address, key Word256) (Word256, error) {
	// Check cache
	cache.RLock()
	info, ok := cache.lookupStorage(addr, key)
	if ok {
		return info.value, nil
	}
	// Get or load storage
	acc, storage, removed, dirty := cache.accounts[addr].unpack()
	if removed {
		return Zero256, fmt.Errorf("GetStorage on a removed account %s", addr)
	}
	cache.RUnlock()
	cache.Lock()
	defer cache.Unlock()

	if acc != nil && storage == nil {
		storage = makeStorage(cache.db, acc.StorageRoot())
		cache.accounts[addr] = accountInfo{acc, storage, false, dirty}
	} else if acc == nil {
		return Zero256, nil
	}
	// Load and set cache
	_, val, _ := storage.Get(key.Bytes())
	value := LeftPadWord256(val)
	cache.setStorage(addr, key, storageInfo{value, false})
	return value, nil
}

// NOTE: Set value to zero to removed from the trie.
func (cache *BlockCache) SetStorage(addr acm.Address, key Word256, value Word256) error {
	cache.Lock()
	defer cache.Unlock()
	_, _, removed, _ := cache.accounts[addr].unpack()
	if removed {
		return fmt.Errorf("SetStorage on a removed account %s", addr)
	}
	cache.setStorage(addr, key, storageInfo{value, true})
	return nil
}

func (cache *BlockCache) IterateStorage(address acm.Address, consumer func(key, value Word256) (stop bool)) (bool, error) {
	cache.RLock()
	defer cache.RUnlock()
	// Try cache first for early exit
	for key, info := range cache.storages[address] {
		if consumer(key, info.value) {
			return true, nil
		}
	}

	return cache.backend.IterateStorage(address, consumer)
}

// BlockCache.storage
//-------------------------------------
// BlockCache.names

func (cache *BlockCache) GetNameRegEntry(name string) *NameRegEntry {
	cache.RLock()
	entry, removed, _ := cache.names[name].unpack()
	cache.RUnlock()
	if removed {
		return nil
	} else if entry != nil {
		return entry
	} else {
		entry = cache.backend.GetNameRegEntry(name)
		cache.Lock()
		cache.names[name] = nameInfo{entry, false, false}
		cache.Unlock()
		return entry
	}
}

func (cache *BlockCache) UpdateNameRegEntry(entry *NameRegEntry) {
	cache.Lock()
	defer cache.Unlock()
	cache.names[entry.Name] = nameInfo{entry, false, true}
}

func (cache *BlockCache) RemoveNameRegEntry(name string) {
	cache.Lock()
	defer cache.Unlock()
	_, removed, _ := cache.names[name].unpack()
	if removed {
		panic("RemoveNameRegEntry on a removed entry")
	}
	cache.names[name] = nameInfo{nil, true, false}
}

// BlockCache.names
//-------------------------------------

// CONTRACT the updates are in deterministic order.
func (cache *BlockCache) Sync() {
	cache.Lock()
	defer cache.Unlock()
	// Determine order for storage updates
	// The address comes first so it'll be grouped.
	storageKeys := make([]Tuple256, 0, len(cache.storages))
	for address, keyInfoMap := range cache.storages {
		for key, _ := range keyInfoMap {
			storageKeys = append(storageKeys, Tuple256{First: address.Word256(), Second: key})
		}
	}
	Tuple256Slice(storageKeys).Sort()

	// Update storage for all account/key.
	// Later we'll iterate over all the users and save storage + update storage root.
	var (
		curAddr       acm.Address
		curAcc        acm.Account
		curAccRemoved bool
		curStorage    merkle.Tree
	)
	for _, storageKey := range storageKeys {
		addrWord256, key := Tuple256Split(storageKey)
		addr := acm.AddressFromWord256(addrWord256)
		if addr != curAddr || curAcc == nil {
			acc, storage, removed, _ := cache.accounts[addr].unpack()
			if !removed && storage == nil {
				storage = makeStorage(cache.db, acc.StorageRoot())
			}
			curAddr = addr
			curAcc = acc
			curAccRemoved = removed
			curStorage = storage
		}
		if curAccRemoved {
			continue
		}
		value, dirty := cache.storages[acm.AddressFromWord256(storageKey.First)][storageKey.Second].unpack()
		if !dirty {
			continue
		}
		if value.IsZero() {
			curStorage.Remove(key.Bytes())
		} else {
			curStorage.Set(key.Bytes(), value.Bytes())
			cache.accounts[addr] = accountInfo{curAcc, curStorage, false, true}
		}
	}

	// Determine order for accounts
	addrs := []acm.Address{}
	for addr := range cache.accounts {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool {
		return addrs[i].String() < addrs[j].String()
	})

	// Update or delete accounts.
	for _, addr := range addrs {
		acc, storage, removed, dirty := cache.accounts[addr].unpack()
		if removed {
			cache.backend.RemoveAccount(addr)
		} else {
			if acc == nil {
				continue
			}
			if storage != nil {
				newStorageRoot := storage.Save()
				if !bytes.Equal(newStorageRoot, acc.StorageRoot()) {
					acc = acm.AsMutableAccount(acc).SetStorageRoot(newStorageRoot)
					dirty = true
				}
			}
			if dirty {
				cache.backend.UpdateAccount(acc)
			}
		}
	}

	// Determine order for names
	// note names may be of any length less than some limit
	nameStrs := []string{}
	for nameStr := range cache.names {
		nameStrs = append(nameStrs, nameStr)
	}
	sort.Strings(nameStrs)

	// Update or delete names.
	for _, nameStr := range nameStrs {
		entry, removed, dirty := cache.names[nameStr].unpack()
		if removed {
			removed := cache.backend.RemoveNameRegEntry(nameStr)
			if !removed {
				panic(fmt.Sprintf("Could not remove namereg entry to be removed: %s", nameStr))
			}
		} else {
			if entry == nil {
				continue
			}
			if dirty {
				cache.backend.UpdateNameRegEntry(entry)
			}
		}
	}
}

func (cache *BlockCache) lookupStorage(address acm.Address, key Word256) (storageInfo, bool) {
	keyInfoMap, ok := cache.storages[address]
	if !ok {
		return storageInfo{}, false
	}
	info, ok := keyInfoMap[key]
	return info, ok
}

func (cache *BlockCache) setStorage(address acm.Address, key Word256, info storageInfo) {
	keyInfoMap, ok := cache.storages[address]
	if !ok {
		keyInfoMap = make(map[Word256]storageInfo)
		cache.storages[address] = keyInfoMap
	}
	keyInfoMap[key] = info
}

//-----------------------------------------------------------------------------

type accountInfo struct {
	account acm.Account
	storage merkle.Tree
	removed bool
	dirty   bool
}

func (accInfo accountInfo) unpack() (acm.Account, merkle.Tree, bool, bool) {
	return accInfo.account, accInfo.storage, accInfo.removed, accInfo.dirty
}

type storageInfo struct {
	value Word256
	dirty bool
}

func (stjInfo storageInfo) unpack() (Word256, bool) {
	return stjInfo.value, stjInfo.dirty
}

type nameInfo struct {
	name    *NameRegEntry
	removed bool
	dirty   bool
}

func (nInfo nameInfo) unpack() (*NameRegEntry, bool, bool) {
	return nInfo.name, nInfo.removed, nInfo.dirty
}
