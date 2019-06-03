// Copyright 2019 Monax Industries Limited
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

package proposal

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"
	"sync"

	"github.com/hyperledger/burrow/txs/payload"
)

// The Cache helps prevent unnecessary IAVLTree updates and garbage generation.
type Cache struct {
	sync.RWMutex
	backend   Reader
	proposals map[[sha256.Size]byte]*proposalInfo
}

type proposalInfo struct {
	sync.RWMutex
	ballot  *payload.Ballot
	removed bool
	updated bool
}

type ProposalHash [sha256.Size]byte

type ProposalHashArray []ProposalHash

func (p ProposalHashArray) Len() int {
	return len(p)
}

func (p ProposalHashArray) Swap(i, j int) {
	p[j], p[i] = p[i], p[j]
}

func (p ProposalHashArray) Less(i, j int) bool {
	switch bytes.Compare(p[i][:], p[j][:]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		panic("bytes.Compare returned invalid value")
	}
}

var _ Writer = &Cache{}

// Returns a Cache that wraps an underlying NameRegCacheGetter to use on a cache miss, can write to an
// output Writer via Sync.
// Not goroutine safe, use syncStateCache if you need concurrent access
func NewCache(backend Reader) *Cache {
	return &Cache{
		backend:   backend,
		proposals: make(map[[sha256.Size]byte]*proposalInfo),
	}
}

func (cache *Cache) GetProposal(proposalHash []byte) (*payload.Ballot, error) {
	proposalInfo, err := cache.get(proposalHash)
	if err != nil {
		return nil, err
	}
	proposalInfo.RLock()
	defer proposalInfo.RUnlock()
	if proposalInfo.removed {
		return nil, nil
	}
	return proposalInfo.ballot, nil
}

func (cache *Cache) UpdateProposal(proposalHash []byte, ballot *payload.Ballot) error {
	proposalInfo, err := cache.get(proposalHash)
	if err != nil {
		return err
	}
	proposalInfo.Lock()
	defer proposalInfo.Unlock()
	if proposalInfo.removed {
		return fmt.Errorf("UpdateProposal on a removed proposal: %x", proposalHash)
	}

	proposalInfo.ballot = ballot
	proposalInfo.updated = true
	return nil
}

func (cache *Cache) RemoveProposal(proposalHash []byte) error {
	proposalInfo, err := cache.get(proposalHash)
	if err != nil {
		return err
	}
	proposalInfo.Lock()
	defer proposalInfo.Unlock()
	if proposalInfo.removed {
		return fmt.Errorf("RemoveProposal on removed proposal: %x", proposalHash)
	}
	proposalInfo.removed = true
	return nil
}

// Writes whatever is in the cache to the output Writer state. Does not flush the cache, to do that call Reset()
// after Sync or use Flusth if your wish to use the output state as your next backend
func (cache *Cache) Sync(state Writer) error {
	cache.Lock()
	defer cache.Unlock()
	// Determine order for names
	// note names may be of any length less than some limit
	var hashes ProposalHashArray
	for hash := range cache.proposals {
		hashes = append(hashes, hash)
	}
	sort.Stable(hashes)

	// Update or delete names.
	for _, hash := range hashes {
		proposalInfo := cache.proposals[hash]
		proposalInfo.RLock()
		if proposalInfo.removed {
			err := state.RemoveProposal(hash[:])
			if err != nil {
				proposalInfo.RUnlock()
				return err
			}
		} else if proposalInfo.updated {
			err := state.UpdateProposal(hash[:], proposalInfo.ballot)
			if err != nil {
				proposalInfo.RUnlock()
				return err
			}
		}
		proposalInfo.RUnlock()
	}
	return nil
}

// Resets the cache to empty initialising the backing map to the same size as the previous iteration.
func (cache *Cache) Reset(backend Reader) {
	cache.Lock()
	defer cache.Unlock()
	cache.backend = backend
	cache.proposals = make(map[[sha256.Size]byte]*proposalInfo)
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

// Get the cache accountInfo item creating it if necessary
func (cache *Cache) get(proposalHash []byte) (*proposalInfo, error) {
	var hash ProposalHash
	copy(hash[:], proposalHash)
	cache.RLock()
	propInfo := cache.proposals[hash]
	cache.RUnlock()
	if propInfo == nil {
		cache.Lock()
		defer cache.Unlock()
		propInfo = cache.proposals[hash]
		if propInfo == nil {
			prop, err := cache.backend.GetProposal(proposalHash)
			if err != nil {
				return nil, err
			}
			propInfo = &proposalInfo{
				ballot: prop,
			}
			cache.proposals[hash] = propInfo
		}
	}
	return propInfo, nil
}
