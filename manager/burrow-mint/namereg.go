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

package burrowmint

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	sm "github.com/monax/burrow/manager/burrow-mint/state"

	core_types "github.com/monax/burrow/core/types"
	event "github.com/monax/burrow/event"
)

// NameReg is part of the pipe for BurrowMint and provides the implementation
// for the pipe to call into the BurrowMint application

type namereg struct {
	burrowMint    *BurrowMint
	filterFactory *event.FilterFactory
}

func newNameReg(burrowMint *BurrowMint) *namereg {

	ff := event.NewFilterFactory()

	ff.RegisterFilterPool("name", &sync.Pool{
		New: func() interface{} {
			return &NameRegNameFilter{}
		},
	})

	ff.RegisterFilterPool("owner", &sync.Pool{
		New: func() interface{} {
			return &NameRegOwnerFilter{}
		},
	})

	ff.RegisterFilterPool("data", &sync.Pool{
		New: func() interface{} {
			return &NameRegDataFilter{}
		},
	})

	ff.RegisterFilterPool("expires", &sync.Pool{
		New: func() interface{} {
			return &NameRegExpiresFilter{}
		},
	})

	return &namereg{burrowMint, ff}
}

func (this *namereg) Entry(key string) (*core_types.NameRegEntry, error) {
	st := this.burrowMint.GetState() // performs a copy
	entry := st.GetNameRegEntry(key)
	if entry == nil {
		return nil, fmt.Errorf("Entry %s not found", key)
	}
	return entry, nil
}

func (this *namereg) Entries(filters []*event.FilterData) (*core_types.ResultListNames, error) {
	var blockHeight int
	var names []*core_types.NameRegEntry
	state := this.burrowMint.GetState()
	blockHeight = state.LastBlockHeight
	filter, err := this.filterFactory.NewFilter(filters)
	if err != nil {
		return nil, fmt.Errorf("Error in query: " + err.Error())
	}
	state.GetNames().Iterate(func(key, value []byte) bool {
		nre := sm.DecodeNameRegEntry(value)
		if filter.Match(nre) {
			names = append(names, nre)
		}
		return false
	})
	return &core_types.ResultListNames{blockHeight, names}, nil
}

type ResultListNames struct {
	BlockHeight int                        `json:"block_height"`
	Names       []*core_types.NameRegEntry `json:"names"`
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
	nre, ok := v.(*core_types.NameRegEntry)
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
		return fmt.Errorf("Wrong value type.")
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
	nre, ok := v.(*core_types.NameRegEntry)
	if !ok {
		return false
	}
	return this.match(nre.Owner, this.value)
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
	nre, ok := v.(*core_types.NameRegEntry)
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
	nre, ok := v.(*core_types.NameRegEntry)
	if !ok {
		return false
	}
	return this.match(int64(nre.Expires), this.value)
}
