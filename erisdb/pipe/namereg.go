package pipe

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	sm "github.com/eris-ltd/eris-db/state"
	"github.com/eris-ltd/eris-db/tmsp"
	"github.com/eris-ltd/eris-db/txs"
)

// The net struct.
type namereg struct {
	erisdbApp     *tmsp.ErisDBApp
	filterFactory *FilterFactory
}

func newNamereg(erisdbApp *tmsp.ErisDBApp) *namereg {

	ff := NewFilterFactory()

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

	return &namereg{erisdbApp, ff}
}

func (this *namereg) Entry(key string) (*types.NameRegEntry, error) {
	st := this.erisdbApp.GetState() // performs a copy
	entry := st.GetNameRegEntry(key)
	if entry == nil {
		return nil, fmt.Errorf("Entry %s not found", key)
	}
	return entry, nil
}

func (this *namereg) Entries(filters []*FilterData) (*ResultListNames, error) {
	var blockHeight int
	var names []*types.NameRegEntry
	state := this.erisdbApp.GetState()
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
	return &ResultListNames{blockHeight, names}, nil
}

type ResultListNames struct {
	BlockHeight int                   `json:"block_height"`
	Names       []*types.NameRegEntry `json:"names"`
}

// Filter for namereg name. This should not be used to get individual entries by name.
// Ops: == or !=
type NameRegNameFilter struct {
	op    string
	value string
	match func(string, string) bool
}

func (this *NameRegNameFilter) Configure(fd *FilterData) error {
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
	nre, ok := v.(*types.NameRegEntry)
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

func (this *NameRegOwnerFilter) Configure(fd *FilterData) error {
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
	nre, ok := v.(*types.NameRegEntry)
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

func (this *NameRegDataFilter) Configure(fd *FilterData) error {
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
	nre, ok := v.(*types.NameRegEntry)
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

func (this *NameRegExpiresFilter) Configure(fd *FilterData) error {
	val, err := ParseNumberValue(fd.Value)
	if err != nil {
		return err
	}
	match, err2 := GetRangeFilter(fd.Op, "expires")
	if err2 != nil {
		return err2
	}
	this.match = match
	this.op = fd.Op
	this.value = val
	return nil
}

func (this *NameRegExpiresFilter) Match(v interface{}) bool {
	nre, ok := v.(*types.NameRegEntry)
	if !ok {
		return false
	}
	return this.match(int64(nre.Expires), this.value)
}
