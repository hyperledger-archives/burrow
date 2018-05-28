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

import acm "github.com/hyperledger/burrow/account"

// NameReg provides a global key value store based on Name, Data pairs that are subject to expiry and ownership by an
// account.
type NameRegEntry struct {
	// registered name for the entry
	Name string
	// address that created the entry
	Owner acm.Address
	// data to store under this name
	Data string
	// block at which this entry expires
	Expires uint64
}

type NameRegGetter interface {
	GetNameRegEntry(name string) (*NameRegEntry, error)
}

type NameRegUpdater interface {
	// Updates the name entry creating it if it does not exist
	UpdateNameRegEntry(entry *NameRegEntry) error
	// Remove the name entry
	RemoveNameRegEntry(name string) error
}

type NameRegWriter interface {
	NameRegGetter
	NameRegUpdater
}

type NameRegIterable interface {
	NameRegGetter
	IterateNameRegEntries(consumer func(*NameRegEntry) (stop bool)) (stopped bool, err error)
}
