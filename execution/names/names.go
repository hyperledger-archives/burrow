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
	"github.com/hyperledger/burrow/crypto"
	"github.com/tendermint/go-amino"
)

var (
	MinNameRegistrationPeriod uint64 = 5

	// NOTE: base costs and validity checks are here so clients
	// can use them without importing state

	// cost for storing a name for a block is
	// CostPerBlock*CostPerByte*(len(data) + 32)
	NameByteCostMultiplier  uint64 = 1
	NameBlockCostMultiplier uint64 = 1

	MaxNameLength = 64
	MaxDataLength = 1 << 16
)

// NameReg provides a global key value store based on Name, Data pairs that are subject to expiry and ownership by an
// account.
type Entry struct {
	// registered name for the entry
	Name string
	// address that created the entry
	Owner crypto.Address
	// data to store under this name
	Data string
	// block at which this entry expires
	Expires uint64
}

var cdc = amino.NewCodec()

func (e *Entry) Encode() ([]byte, error) {
	return cdc.MarshalBinary(e)
}

func DecodeEntry(entryBytes []byte) (*Entry, error) {
	entry := new(Entry)
	err := cdc.UnmarshalBinary(entryBytes, entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

type Getter interface {
	GetNameEntry(name string) (*Entry, error)
}

type Updater interface {
	// Updates the name entry creating it if it does not exist
	UpdateNameEntry(entry *Entry) error
	// Remove the name entry
	RemoveNameEntry(name string) error
}

type Writer interface {
	Getter
	Updater
}

type Iterable interface {
	Getter
	IterateNameEntries(consumer func(*Entry) (stop bool)) (stopped bool, err error)
}

// base cost is "effective" number of bytes
func NameBaseCost(name, data string) uint64 {
	return uint64(len(data) + 32)
}

func NameCostPerBlock(baseCost uint64) uint64 {
	return NameBlockCostMultiplier * NameByteCostMultiplier * baseCost
}
