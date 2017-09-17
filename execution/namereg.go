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
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/blockchain"
	event "github.com/hyperledger/burrow/event"
)

// NameReg is part of the pipe for BurrowMint and provides the implementation
// for the pipe to call into the BurrowMint application

type namereg struct {
	state         *State
	blockchain    blockchain.Blockchain
	filterFactory *event.FilterFactory
}

type NameRegEntry struct {
	Name    string          `json:"name"`    // registered name for the entry
	Owner   account.Address `json:"owner"`   // address that created the entry
	Data    string          `json:"data"`    // data to store under this name
	Expires uint64          `json:"expires"` // block at which this entry expires
}

func newNameReg(state *State, blockchain blockchain.Blockchain) *namereg {
	filterFactory := event.NewFilterFactory()

	filterFactory.RegisterFilterPool("name", &sync.Pool{
		New: func() interface{} {
			return &NameRegNameFilter{}
		},
	})

	filterFactory.RegisterFilterPool("owner", &sync.Pool{
		New: func() interface{} {
			return &NameRegOwnerFilter{}
		},
	})

	filterFactory.RegisterFilterPool("data", &sync.Pool{
		New: func() interface{} {
			return &NameRegDataFilter{}
		},
	})

	filterFactory.RegisterFilterPool("expires", &sync.Pool{
		New: func() interface{} {
			return &NameRegExpiresFilter{}
		},
	})

	return &namereg{
		state:         state,
		blockchain:    blockchain,
		filterFactory: filterFactory,
	}
}

func (nr *namereg) Entry(key string) (*NameRegEntry, error) {
	entry := nr.state.GetNameRegEntry(key)
	if entry == nil {
		return nil, fmt.Errorf("entry %s not found", key)
	}
	return entry, nil
}

func (nr *namereg) Entries(filters []*event.FilterData) (*ResultListNames, error) {
	var names []*NameRegEntry
	blockHeight := nr.blockchain.LastBlockHeight()
	filter, err := nr.filterFactory.NewFilter(filters)
	if err != nil {
		return nil, fmt.Errorf("Error in query: " + err.Error())
	}
	nr.state.GetNames().Iterate(func(key, value []byte) bool {
		nre := DecodeNameRegEntry(value)
		if filter.Match(nre) {
			names = append(names, nre)
		}
		return false
	})
	return &ResultListNames{blockHeight, names}, nil
}

type ResultListNames struct {
	BlockHeight uint64          `json:"block_height"`
	Names       []*NameRegEntry `json:"names"`
}

// Filter for namereg name. This should not be used to get individual entries by name.
// Ops: == or !=
type NameRegNameFilter struct {
	op    string
	value string
	match func(string, string) bool
}

func (this *NameRegNameFilter) Configure(fd *event.FilterData) error {
	op := fd.Op
	val := fd.Value

	if op == "==" {
		this.match = func(a, b string) bool {
			return a == b
		}
	} else if op == "!=" {
		this.match = func(a, b string) bool {
			return a != b
		}
	} else {
		return fmt.Errorf("Op: " + this.op + " is not supported for 'name' filtering")
	}
	this.op = op
	this.value = val
	return nil
}

func (this *NameRegNameFilter) Match(v interface{}) bool {
	nre, ok := v.(*NameRegEntry)
	if !ok {
		return false
	}
	return this.match(nre.Name, this.value)
}

// Filter for owner.
// Ops: == or !=
type NameRegOwnerFilter struct {
	op    string
	value []byte
	match func([]byte, []byte) bool
}

func (this *NameRegOwnerFilter) Configure(fd *event.FilterData) error {
	op := fd.Op
	val, err := hex.DecodeString(fd.Value)

	if err != nil {
		return fmt.Errorf("wrong value type.")
	}
	if op == "==" {
		this.match = func(a, b []byte) bool {
			return bytes.Equal(a, b)
		}
	} else if op == "!=" {
		this.match = func(a, b []byte) bool {
			return !bytes.Equal(a, b)
		}
	} else {
		return fmt.Errorf("Op: " + this.op + " is not supported for 'owner' filtering")
	}
	this.op = op
	this.value = val
	return nil
}

func (this *NameRegOwnerFilter) Match(v interface{}) bool {
	nre, ok := v.(*NameRegEntry)
	if !ok {
		return false
	}
	return this.match(nre.Owner.Bytes(), this.value)
}

// Filter for namereg data. Useful for example if you store an ipfs hash and know the hash but need the key.
// Ops: == or !=
type NameRegDataFilter struct {
	op    string
	value string
	match func(string, string) bool
}

func (this *NameRegDataFilter) Configure(fd *event.FilterData) error {
	op := fd.Op
	val := fd.Value

	if op == "==" {
		this.match = func(a, b string) bool {
			return a == b
		}
	} else if op == "!=" {
		this.match = func(a, b string) bool {
			return a != b
		}
	} else {
		return fmt.Errorf("Op: " + this.op + " is not supported for 'data' filtering")
	}
	this.op = op
	this.value = val
	return nil
}

func (this *NameRegDataFilter) Match(v interface{}) bool {
	nre, ok := v.(*NameRegEntry)
	if !ok {
		return false
	}
	return this.match(nre.Data, this.value)
}

// Filter for expires.
// Ops: All
type NameRegExpiresFilter struct {
	op    string
	value int64
	match func(int64, int64) bool
}

func (this *NameRegExpiresFilter) Configure(fd *event.FilterData) error {
	val, err := event.ParseNumberValue(fd.Value)
	if err != nil {
		return err
	}
	match, err2 := event.GetRangeFilter(fd.Op, "expires")
	if err2 != nil {
		return err2
	}
	this.match = match
	this.op = fd.Op
	this.value = val
	return nil
}

func (this *NameRegExpiresFilter) Match(v interface{}) bool {
	nre, ok := v.(*NameRegEntry)
	if !ok {
		return false
	}
	return this.match(int64(nre.Expires), this.value)
}
