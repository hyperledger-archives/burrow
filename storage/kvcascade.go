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

type KVCascade []KVIterableReader

func (kvc KVCascade) Get(key []byte) []byte {
	for _, kvs := range kvc {
		value := kvs.Get(key)
		if value != nil {
			return value
		}
	}
	return nil
}

func (kvc KVCascade) Has(key []byte) bool {
	for _, kvs := range kvc {
		has := kvs.Has(key)
		if has {
			return true
		}
	}
	return false
}

func (kvc KVCascade) Iterator(low, high []byte) KVIterator {
	iterators := make([]KVIterator, len(kvc))
	for i, kvs := range kvc {
		iterators[i] = kvs.Iterator(low, high)
	}
	return NewMultiIterator(false, iterators...)
}

func (kvc KVCascade) ReverseIterator(low, high []byte) KVIterator {
	iterators := make([]KVIterator, len(kvc))
	for i, kvs := range kvc {
		iterators[i] = kvs.ReverseIterator(low, high)
	}
	return NewMultiIterator(true, iterators...)
}
