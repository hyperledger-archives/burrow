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
	"fmt"
	"sync"

	"github.com/hyperledger/burrow/crypto"
)

// Cache helps prevent unnecessary IAVLTree updates and garbage generation.
type Cache struct {
	sync.RWMutex
	backend  Reader
	registry map[crypto.Address]*nodeInfo
}

type nodeInfo struct {
	sync.RWMutex
	node    *NodeIdentity
	removed bool
	updated bool
}

var _ Writer = &Cache{}

// NewCache returns a Cache which can write to an output Writer via Sync.
// Not goroutine safe, use syncStateCache if you need concurrent access
func NewCache(backend Reader) *Cache {
	return &Cache{
		backend:  backend,
		registry: make(map[crypto.Address]*nodeInfo),
	}
}

func (cache *Cache) GetNode(addr crypto.Address) (*NodeIdentity, error) {
	info, err := cache.get(addr)
	if err != nil {
		return nil, err
	}
	info.RLock()
	defer info.RUnlock()
	if info.removed {
		return nil, nil
	}
	return info.node, nil
}

func (cache *Cache) GetNodes() NodeList {
	nodes := make(NodeList)
	for addr := range cache.registry {
		n, err := cache.GetNode(addr)
		if err != nil {
			continue
		}
		nodes[addr] = n
	}
	return nodes
}

func (cache *Cache) UpdateNode(addr crypto.Address, node *NodeIdentity) error {
	info, err := cache.get(addr)
	if err != nil {
		return err
	}
	info.Lock()
	defer info.Unlock()
	if info.removed {
		return fmt.Errorf("UpdateNode on a removed node: %x", addr)
	}

	info.node = node
	info.updated = true
	return nil
}

func (cache *Cache) RemoveNode(addr crypto.Address) error {
	info, err := cache.get(addr)
	if err != nil {
		return err
	}
	info.Lock()
	defer info.Unlock()
	if info.removed {
		return fmt.Errorf("RemoveNode on removed node: %x", addr)
	}
	info.removed = true
	return nil
}

// Sync writes whatever is in the cache to the output state. Does not flush the cache, to do that call Reset()
// after Sync or use Flush if your wish to use the output state as your next backend
func (cache *Cache) Sync(state Writer) error {
	cache.Lock()
	defer cache.Unlock()
	var addresses []crypto.Address
	for addr := range cache.registry {
		addresses = append(addresses, addr)
	}

	for _, addr := range addresses {
		info := cache.registry[addr]
		info.RLock()
		if info.removed {
			err := state.RemoveNode(addr)
			if err != nil {
				info.RUnlock()
				return err
			}
		} else if info.updated {
			err := state.UpdateNode(addr, info.node)
			if err != nil {
				info.RUnlock()
				return err
			}
		}
		info.RUnlock()
	}
	return nil
}

// Reset the cache to empty initialising the backing map to the same size as the previous iteration
func (cache *Cache) Reset(backend Reader) {
	cache.Lock()
	defer cache.Unlock()
	cache.backend = backend
	cache.registry = make(map[crypto.Address]*nodeInfo)
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

func (cache *Cache) get(addr crypto.Address) (*nodeInfo, error) {
	cache.RLock()
	info := cache.registry[addr]
	cache.RUnlock()
	if info == nil {
		cache.Lock()
		defer cache.Unlock()
		info = cache.registry[addr]
		if info == nil {
			node, err := cache.backend.GetNode(addr)
			if err != nil {
				return nil, err
			}
			info = &nodeInfo{
				node: node,
			}
			cache.registry[addr] = info
		}
	}
	return info, nil
}
