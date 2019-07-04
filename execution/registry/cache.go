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

package registry

import (
	"sync"

	"github.com/hyperledger/burrow/crypto"
)

// Cache helps prevent unnecessary IAVLTree updates and garbage generation.
type Cache struct {
	sync.RWMutex
	backend  Reader
	registry map[crypto.Address]*RegisteredNode
}

type nodeInfo struct {
	sync.RWMutex
}

var _ Writer = &Cache{}

// Returns a Cache, can write to an output Writer via Sync.
// Not goroutine safe, use syncStateCache if you need concurrent access
func NewCache(backend Reader) *Cache {
	return &Cache{
		backend: backend,
	}
}

func (cache *Cache) RegisterNode(val crypto.Address, regNode *RegisteredNode) error {
	return nil
}

func (cache *Cache) GetNetworkRegistry() (map[crypto.Address]*RegisteredNode, error) {
	return nil, nil
}

// Writes whatever is in the cache to the output Writer state. Does not flush the cache, to do that call Reset()
// after Sync or use Flush if your wish to use the output state as your next backend
func (cache *Cache) Sync(state Writer) error {
	cache.Lock()
	defer cache.Unlock()
	return nil
}

// Resets the cache to empty initialising the backing map to the same size as the previous iteration.
func (cache *Cache) Reset(backend Reader) {
	cache.Lock()
	defer cache.Unlock()
}

// Syncs the Cache and Resets it to use Writer as the backend Reader
func (cache *Cache) Flush(output Writer, backend Reader) error {
	err := cache.Sync(output)
	if err != nil {
		return err
	}
	cache.Reset(backend)
	return nil
}

func (cache *Cache) Backend() Reader {
	return cache.backend
}

func (cache *Cache) get(val crypto.Address) (*RegisteredNode, error) {
	cache.RLock()
	defer cache.RUnlock()
	return nil, nil
}
