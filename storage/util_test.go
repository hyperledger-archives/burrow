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
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sendKVPair(ch chan<- KVPair, kvs []KVPair) {
	for _, kv := range kvs {
		ch <- kv
	}
	close(ch)
}

func collectIterator(it KVIterator) KVPairs {
	var kvp []KVPair
	for it.Valid() {
		kvp = append(kvp, KVPair{it.Key(), it.Value()})
		it.Next()
	}
	return kvp
}

func kvPairs(kvs ...string) KVPairs {
	n := len(kvs) / 2
	kvp := make([]KVPair, 0, n)
	for i := 0; i < 2*n; i += 2 {
		kvp = append(kvp, KVPair{[]byte(kvs[i]), []byte(kvs[i+1])})
	}
	return kvp
}

func assertIteratorSorted(t *testing.T, it KVIterator, reverse bool) {
	prev := ""
	for it.Valid() {
		strKey := string(it.Key())
		t.Log(strKey, "=>", string(it.Value()))
		if prev == "" {
			prev = strKey
		}
		// Assert non-decreasing sequence of keys
		if reverse {
			assert.False(t, strings.Compare(prev, strKey) == -1)
		} else {
			assert.False(t, strings.Compare(prev, strKey) == 1)
		}
		prev = strKey
		it.Next()
	}
}

func iteratorOver(kvp []KVPair, reverse ...bool) *ChannelIterator {
	var sortable sort.Interface = KVPairs(kvp)
	if len(reverse) > 0 && reverse[0] {
		sortable = sort.Reverse(sortable)
	}
	sort.Stable(sortable)
	ch := make(chan KVPair)
	var start, end []byte
	if len(kvp) > 0 {
		start, end = kvp[0].Key, kvp[len(kvp)-1].Key
	}
	go sendKVPair(ch, kvp)
	ci := NewChannelIterator(ch, start, end)
	return ci
}

func bz(s string) []byte {
	return []byte(s)
}
