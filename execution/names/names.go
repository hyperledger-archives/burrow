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

	"github.com/hyperledger/burrow/event/query"
	"github.com/tendermint/go-amino"
)

var MinNameRegistrationPeriod uint64 = 5

const (

	// NOTE: base costs and validity checks are here so clients
	// can use them without importing state

	// cost for storing a name for a block is
	// CostPerBlock*CostPerByte*(len(data) + 32)
	NameByteCostMultiplier  uint64 = 1
	NameBlockCostMultiplier uint64 = 1

	MaxNameLength = 64
	MaxDataLength = 1 << 16
)

var cdc = amino.NewCodec()

func (e *Entry) Encode() ([]byte, error) {
	return cdc.MarshalBinary(e)
}

func (e *Entry) String() string {
	return fmt.Sprintf("NameEntry{%v -> %v; Expires: %v, Owner: %v}", e.Name, e.Data, e.Expires, e.Owner)
}

type TaggedEntry struct {
	*Entry
	query.Tagged
}

func (e *Entry) Tagged() *TaggedEntry {
	return &TaggedEntry{
		Entry:  e,
		Tagged: query.MustReflectTags(e),
	}
}

func DecodeEntry(entryBytes []byte) (*Entry, error) {
	entry := new(Entry)
	err := cdc.UnmarshalBinary(entryBytes, entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

type Reader interface {
	GetName(name string) (*Entry, error)
}

type Writer interface {
	// Updates the name entry creating it if it does not exist
	UpdateName(entry *Entry) error
	// Remove the name entry
	RemoveName(name string) error
}

type ReaderWriter interface {
	Reader
	Writer
}

type Iterable interface {
	IterateNames(consumer func(*Entry) (stop bool)) (stopped bool, err error)
}

type IterableReader interface {
	Iterable
	Reader
}

type IterableReaderWriter interface {
	Iterable
	ReaderWriter
}

// base cost is "effective" number of bytes
func NameBaseCost(name, data string) uint64 {
	return uint64(len(data) + 32)
}

func NameCostPerBlock(baseCost uint64) uint64 {
	return NameBlockCostMultiplier * NameByteCostMultiplier * baseCost
}

func NameCostForExpiryIn(name, data string, expiresIn uint64) uint64 {
	return NameCostPerBlock(NameBaseCost(name, data)) * expiresIn
}
