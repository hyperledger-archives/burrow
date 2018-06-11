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

package names

import (
	"fmt"
	"sort"
	"sync"
)

// The NameRegCache helps prevent unnecessary IAVLTree updates and garbage generation.
type NameRegCache struct {
	sync.RWMutex
	backend NameRegGetter
	names   map[string]*nameInfo
}

type nameInfo struct {
	sync.RWMutex
	entry   *NameRegEntry
	removed bool
	updated bool
}

var _ NameRegWriter = &NameRegCache{}

// Returns a NameRegCache that wraps an underlying NameRegCacheGetter to use on a cache miss, can write to an
// output NameRegWriter via Sync.
// Not goroutine safe, use syncStateCache if you need concurrent access
func NewNameRegCache(backend NameRegGetter) *NameRegCache {
	return &NameRegCache{
		backend: backend,
		names:   make(map[string]*nameInfo),
	}
}

func (cache *NameRegCache) GetNameRegEntry(name string) (*NameRegEntry, error) {
	nameInfo, err := cache.get(name)
	if err != nil {
		return nil, err
	}
	nameInfo.RLock()
	defer nameInfo.RUnlock()
	if nameInfo.removed {
		return nil, nil
	}
	return nameInfo.entry, nil
}

func (cache *NameRegCache) UpdateNameRegEntry(entry *NameRegEntry) error {
	nameInfo, err := cache.get(entry.Name)
	if err != nil {
		return err
	}
	nameInfo.Lock()
	defer nameInfo.Unlock()
	if nameInfo.removed {
		return fmt.Errorf("UpdateNameRegEntry on a removed name: %s", nameInfo.entry.Name)
	}

	nameInfo.entry = entry
	nameInfo.updated = true
	return nil
}

func (cache *NameRegCache) RemoveNameRegEntry(name string) error {
	nameInfo, err := cache.get(name)
	if err != nil {
		return err
	}
	nameInfo.Lock()
	defer nameInfo.Unlock()
	if nameInfo.removed {
		return fmt.Errorf("RemoveNameRegEntry on removed name: %s", name)
	}
	nameInfo.removed = true
	return nil
}

// Writes whatever is in the cache to the output NameRegWriter state. Does not flush the cache, to do that call Reset()
// after Sync or use Flusth if your wish to use the output state as your next backend
func (cache *NameRegCache) Sync(state NameRegWriter) error {
	cache.Lock()
	defer cache.Unlock()
	// Determine order for names
	// note names may be of any length less than some limit
	var names sort.StringSlice
	for nameStr := range cache.names {
		names = append(names, nameStr)
	}
	sort.Stable(names)

	// Update or delete names.
	for _, name := range names {
		nameInfo := cache.names[name]
		nameInfo.RLock()
		if nameInfo.removed {
			err := state.RemoveNameRegEntry(name)
			if err != nil {
				nameInfo.RUnlock()
				return err
			}
		} else if nameInfo.updated {
			err := state.UpdateNameRegEntry(nameInfo.entry)
			if err != nil {
				nameInfo.RUnlock()
				return err
			}
		}
		nameInfo.RUnlock()
	}
	return nil
}

// Resets the cache to empty initialising the backing map to the same size as the previous iteration.
func (cache *NameRegCache) Reset(backend NameRegGetter) {
	cache.Lock()
	defer cache.Unlock()
	cache.backend = backend
	cache.names = make(map[string]*nameInfo)
}

// Syncs the NameRegCache and Resets it to use NameRegWriter as the backend NameRegGetter
func (cache *NameRegCache) Flush(state NameRegWriter) error {
	err := cache.Sync(state)
	if err != nil {
		return err
	}
	cache.Reset(state)
	return nil
}

func (cache *NameRegCache) Backend() NameRegGetter {
	return cache.backend
}

// Get the cache accountInfo item creating it if necessary
func (cache *NameRegCache) get(name string) (*nameInfo, error) {
	cache.RLock()
	nmeInfo := cache.names[name]
	cache.RUnlock()
	if nmeInfo == nil {
		cache.Lock()
		defer cache.Unlock()
		nmeInfo = cache.names[name]
		if nmeInfo == nil {
			entry, err := cache.backend.GetNameRegEntry(name)
			if err != nil {
				return nil, err
			}
			nmeInfo = &nameInfo{
				entry: entry,
			}
			cache.names[name] = nmeInfo
		}
	}
	return nmeInfo, nil
}
