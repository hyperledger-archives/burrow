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
	. "github.com/hyperledger/burrow/word"

	"github.com/tendermint/merkleeyes/iavl"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/merkle"
)

func makeStorage(db dbm.DB, root []byte) merkle.Tree {
	storage := iavl.NewIAVLTree(1024, db)
	storage.Load(root)
	return storage
}

// Implements all of evm.State except acm.Creator (it doesn't need to)
var _ acm.UpdaterAndStorage = &BlockCache{}

// The blockcache helps prevent unnecessary IAVLTree updates and garbage generation.
type BlockCache struct {
	db       dbm.DB
	backend  *State
	accounts map[acm.Address]accountInfo
	storages map[Tuple256]storageInfo
	names    map[string]nameInfo
}

func NewBlockCache(backend *State) *BlockCache {
	return &BlockCache{
		// TODO: This is bad and probably the cause of various panics
		db:       backend.db,
		backend:  backend,
		accounts: make(map[acm.Address]accountInfo),
		storages: make(map[Tuple256]storageInfo),
		names:    make(map[string]nameInfo),
	}
}

func (cache *BlockCache) State() *State {
	return cache.backend
}

//-------------------------------------
// BlockCache.account

func (cache *BlockCache) GetAccount(addr acm.Address) *acm.ConcreteAccount {
	acc, _, removed, _ := cache.accounts[addr].unpack()
	if removed {
		return nil
	} else if acc != nil {
		return acc
	} else {
		acc = cache.backend.GetAccount(addr)
		cache.accounts[addr] = accountInfo{acc, nil, false, false}
		return acc
	}
}

func (cache *BlockCache) UpdateAccount(acc *acm.ConcreteAccount) {
	addr := acc.Address
	_, storage, removed, _ := cache.accounts[addr].unpack()
	if removed {
		panic("UpdateAccount on a removed account")
	}
	cache.accounts[addr] = accountInfo{acc, storage, false, true}
}

func (cache *BlockCache) RemoveAccount(addr acm.Address) {
	_, _, removed, _ := cache.accounts[addr].unpack()
	if removed {
		panic("RemoveAccount on a removed account")
	}
	cache.accounts[addr] = accountInfo{nil, nil, true, false}
}

// BlockCache.account
//-------------------------------------
// BlockCache.storage

func (cache *BlockCache) GetStorage(addr acm.Address, key Word256) (value Word256) {
	// Check cache
	info, ok := cache.storages[Tuple256{addr.Word256(), key}]
	if ok {
		return info.value
	}

	// Get or load storage
	acc, storage, removed, dirty := cache.accounts[addr].unpack()
	if removed {
		panic("GetStorage() on removed account")
	}
	if acc != nil && storage == nil {
		storage = makeStorage(cache.db, acc.StorageRoot)
		cache.accounts[addr] = accountInfo{acc, storage, false, dirty}
	} else if acc == nil {
		return Zero256
	}

	// Load and set cache
	_, val_, _ := storage.Get(key.Bytes())
	if val_ != nil {
		value = LeftPadWord256(val_)
	}
	cache.storages[Tuple256{addr.Word256(), key}] = storageInfo{value, false}
	return value
}

// NOTE: Set value to zero to removed from the trie.
func (cache *BlockCache) SetStorage(addr acm.Address, key Word256, value Word256) {
	_, _, removed, _ := cache.accounts[addr].unpack()
	if removed {
		panic("SetStorage() on a removed account")
	}
	cache.storages[Tuple256{addr.Word256(), key}] = storageInfo{value, true}
}

// BlockCache.storage
//-------------------------------------
// BlockCache.names

func (cache *BlockCache) GetNameRegEntry(name string) *NameRegEntry {
	entry, removed, _ := cache.names[name].unpack()
	if removed {
		return nil
	} else if entry != nil {
		return entry
	} else {
		entry = cache.backend.GetNameRegEntry(name)
		cache.names[name] = nameInfo{entry, false, false}
		return entry
	}
}

func (cache *BlockCache) UpdateNameRegEntry(entry *NameRegEntry) {
	name := entry.Name
	cache.names[name] = nameInfo{entry, false, true}
}

func (cache *BlockCache) RemoveNameRegEntry(name string) {
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

	// Determine order for storage updates
	// The address comes first so it'll be grouped.
	storageKeys := make([]Tuple256, 0, len(cache.storages))
	for keyTuple := range cache.storages {
		storageKeys = append(storageKeys, keyTuple)
	}
	Tuple256Slice(storageKeys).Sort()

	// Update storage for all account/key.
	// Later we'll iterate over all the users and save storage + update storage root.
	var (
		curAddr       acm.Address
		curAcc        *acm.ConcreteAccount
		curAccRemoved bool
		curStorage    merkle.Tree
	)
	for _, storageKey := range storageKeys {
		addrWord256, key := Tuple256Split(storageKey)
		addr := acm.AddressFromWord256(addrWord256)
		if addr != curAddr || curAcc == nil {
			acc, storage, removed, _ := cache.accounts[addr].unpack()
			if !removed && storage == nil {
				storage = makeStorage(cache.db, acc.StorageRoot)
			}
			curAddr = addr
			curAcc = acc
			curAccRemoved = removed
			curStorage = storage
		}
		if curAccRemoved {
			continue
		}
		value, dirty := cache.storages[storageKey].unpack()
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
				if !bytes.Equal(newStorageRoot, acc.StorageRoot) {
					acc.StorageRoot = newStorageRoot
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

//-----------------------------------------------------------------------------

type accountInfo struct {
	account *acm.ConcreteAccount
	storage merkle.Tree
	removed bool
	dirty   bool
}

func (accInfo accountInfo) unpack() (*acm.ConcreteAccount, merkle.Tree, bool, bool) {
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
