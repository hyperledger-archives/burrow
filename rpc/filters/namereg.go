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

package filters

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/hyperledger/burrow/execution/names"
)

func NewNameRegFilterFactory() *FilterFactory {
	filterFactory := NewFilterFactory()

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

	return filterFactory
}

// Filter for namereg name. This should not be used to get individual entries by name.
// Ops: == or !=
type NameRegNameFilter struct {
	op    string
	value string
	match func(string, string) bool
}

func (nrnf *NameRegNameFilter) Configure(fd *FilterData) error {
	op := fd.Op
	val := fd.Value

	if op == "==" {
		nrnf.match = func(a, b string) bool {
			return a == b
		}
	} else if op == "!=" {
		nrnf.match = func(a, b string) bool {
			return a != b
		}
	} else {
		return fmt.Errorf("Op: " + nrnf.op + " is not supported for 'name' filtering")
	}
	nrnf.op = op
	nrnf.value = val
	return nil
}

func (nrnf *NameRegNameFilter) Match(v interface{}) bool {
	nre, ok := v.(*names.NameRegEntry)
	if !ok {
		return false
	}
	return nrnf.match(nre.Name, nrnf.value)
}

// Filter for owner.
// Ops: == or !=
type NameRegOwnerFilter struct {
	op    string
	value []byte
	match func([]byte, []byte) bool
}

func (nrof *NameRegOwnerFilter) Configure(fd *FilterData) error {
	op := fd.Op
	val, err := hex.DecodeString(fd.Value)

	if err != nil {
		return fmt.Errorf("wrong value type.")
	}
	if op == "==" {
		nrof.match = func(a, b []byte) bool {
			return bytes.Equal(a, b)
		}
	} else if op == "!=" {
		nrof.match = func(a, b []byte) bool {
			return !bytes.Equal(a, b)
		}
	} else {
		return fmt.Errorf("Op: " + nrof.op + " is not supported for 'owner' filtering")
	}
	nrof.op = op
	nrof.value = val
	return nil
}

func (nrof *NameRegOwnerFilter) Match(v interface{}) bool {
	nre, ok := v.(*names.NameRegEntry)
	if !ok {
		return false
	}
	return nrof.match(nre.Owner.Bytes(), nrof.value)
}

// Filter for namereg data. Useful for example if you store an ipfs hash and know the hash but need the key.
// Ops: == or !=
type NameRegDataFilter struct {
	op    string
	value string
	match func(string, string) bool
}

func (nrdf *NameRegDataFilter) Configure(fd *FilterData) error {
	op := fd.Op
	val := fd.Value

	if op == "==" {
		nrdf.match = func(a, b string) bool {
			return a == b
		}
	} else if op == "!=" {
		nrdf.match = func(a, b string) bool {
			return a != b
		}
	} else {
		return fmt.Errorf("Op: " + nrdf.op + " is not supported for 'data' filtering")
	}
	nrdf.op = op
	nrdf.value = val
	return nil
}

func (nrdf *NameRegDataFilter) Match(v interface{}) bool {
	nre, ok := v.(*names.NameRegEntry)
	if !ok {
		return false
	}
	return nrdf.match(nre.Data, nrdf.value)
}

// Filter for expires.
// Ops: All
type NameRegExpiresFilter struct {
	op    string
	value uint64
	match func(uint64, uint64) bool
}

func (nref *NameRegExpiresFilter) Configure(fd *FilterData) error {
	val, err := ParseNumberValue(fd.Value)
	if err != nil {
		return err
	}
	match, err2 := GetRangeFilter(fd.Op, "expires")
	if err2 != nil {
		return err2
	}
	nref.match = match
	nref.op = fd.Op
	nref.value = val
	return nil
}

func (nref *NameRegExpiresFilter) Match(v interface{}) bool {
	nre, ok := v.(*names.NameRegEntry)
	if !ok {
		return false
	}
	return nref.match(nre.Expires, nref.value)
}
