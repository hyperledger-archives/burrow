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

package storage

import (
	"crypto/sha256"

	"fmt"

	dbm "github.com/tendermint/tendermint/libs/db"
)

type ContentAddressedStore struct {
	db dbm.DB
}

func NewContentAddressedStore(db dbm.DB) *ContentAddressedStore {
	return &ContentAddressedStore{
		db: db,
	}
}

// These function match those used in Hoard

// Put data in the database by saving data with a key that is its sha256 hash
func (cas *ContentAddressedStore) Put(data []byte) ([]byte, error) {
	hasher := sha256.New()
	_, err := hasher.Write(data)
	if err != nil {
		return nil, fmt.Errorf("ContentAddressedStore could not hash data: %v", err)
	}
	hash := hasher.Sum(nil)
	cas.db.SetSync(hash, data)
	return hash, nil
}

func (cas *ContentAddressedStore) Get(hash []byte) ([]byte, error) {
	return cas.db.Get(hash), nil
}
