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

import "github.com/tendermint/iavl"

// We wrap IAVL's tree types in order to provide iteration helpers and to harmonise other interface types with what we
// expect

type ImmutableTree struct {
	*iavl.ImmutableTree
}

func (imt *ImmutableTree) Get(key []byte) []byte {
	_, value := imt.ImmutableTree.Get(key)
	return value
}

func (imt *ImmutableTree) Iterate(start, end []byte, ascending bool, fn func(key []byte, value []byte) error) error {
	var err error
	imt.ImmutableTree.IterateRange(start, end, ascending, func(key, value []byte) bool {
		err = fn(key, value)
		if err != nil {
			// stop
			return true
		}
		return false
	})
	return err
}
