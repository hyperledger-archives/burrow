// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

// NameReg is part of the pipe for ErisMint and provides the implementation
// for the pipe to call into the ErisMint application
package erismint

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	sm "github.com/eris-ltd/eris-db/manager/eris-mint/state"

	core_types "github.com/eris-ltd/eris-db/core/types"
	event "github.com/eris-ltd/eris-db/event"
)

// The net struct.
type namereg struct {
	erisMint      *ErisMint
	filterFactory *event.FilterFactory
}

func newNameReg(erisMint *ErisMint) *namereg {

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

	return &namereg{erisMint, ff}
}

func (this *namereg) Entry(key string) (*core_types.NameRegEntry, error) {
	st := this.erisMint.GetState() // performs a copy
	entry := st.GetNameRegEntry(key)
	if entry == nil {
		return nil, fmt.Errorf("Entry %s not found", key)
	}
	return entry, nil
}

func (this *namereg) Entries(filters []*event.FilterData) (*core_types.ResultListNames, error) {
	var blockHeight int
	var names []*core_types.NameRegEntry
	state := this.erisMint.GetState()
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
