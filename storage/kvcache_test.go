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
	"math/rand"
	"testing"

	"github.com/hyperledger/burrow/binary"

	"github.com/stretchr/testify/assert"
)

func TestKVCache_Iterator(t *testing.T) {
	kvc := NewKVCache()
	kvp := kvPairs(
		"f", "oo",
		"b", "ar",
		"aa", "ooh",
		"im", "aginative",
	)
	sortedKVP := kvPairs(
		"aa", "ooh",
		"b", "ar",
		"f", "oo",
		"im", "aginative",
	)
	for _, kv := range kvp {
		kvc.Set(kv.Key, kv.Value)
	}
	assert.Equal(t, sortedKVP, collectIterator(kvc.Iterator(nil, nil)))
}

func TestKVCache_Iterator2(t *testing.T) {
	assert.Equal(t, []string{"b"}, testIterate(bz("b"), bz("c"), false, "a", "b", "c", "d"))
	assert.Equal(t, []string{"b", "c"}, testIterate(bz("b"), bz("cc"), false, "a", "b", "c", "d"))
	assert.Equal(t, []string{"a", "b", "c", "d"}, testIterate(bz(""), nil, false, "a", "b", "c", "d"))
	assert.Equal(t, []string{"d", "c", "b", "a"}, testIterate(bz(""), nil, true, "a", "b", "c", "d"))
	assert.Equal(t, []string{"a", "b", "c", "d"}, testIterate(nil, nil, false, "a", "b", "c", "d"))

	assert.Equal(t, []string{}, testIterate(bz(""), bz(""), false, "a", "b", "c", "d"))
	assert.Equal(t, []string{}, testIterate(bz("ab"), bz("ab"), false, "a", "b", "c", "d"))
	assert.Equal(t, []string{"a"}, testIterate(bz("0"), bz("ab"), true, "a", "b", "c", "d"))
	assert.Equal(t, []string{"c", "b", "a"}, testIterate(bz("a"), bz("c1"), true, "a", "b", "c", "d"))
	assert.Equal(t, []string{"b", "a"}, testIterate(bz("a"), bz("c"), true, "a", "b", "c", "d"))
	assert.Equal(t, []string{"b", "a"}, testIterate(bz("a"), bz("c"), true, "a", "b", "c", "d"))
	assert.Equal(t, []string{}, testIterate(bz("c"), bz("e"), true, "a", "b"))
	assert.Equal(t, []string{}, testIterate(bz("c"), bz("e"), true, "z", "f"))
}

func BenchmarkKVCache_Iterator_1E6_Inserts(b *testing.B) {
	benchmarkKVCache_Iterator(b, 1E6)
}

func BenchmarkKVCache_Iterator_1E7_Inserts(b *testing.B) {
	benchmarkKVCache_Iterator(b, 1E7)
}

func benchmarkKVCache_Iterator(b *testing.B, inserts int) {
	b.StopTimer()
	cache := NewKVCache()
	rnd := rand.NewSource(23425)
	keyvals := make([][]byte, inserts)
	for i := 0; i < inserts; i++ {
		bs := make([]byte, 8)
		binary.PutInt64BE(bs, rnd.Int63())
		keyvals[i] = bs
	}
	for i := 0; i < inserts; i++ {
		cache.Set(keyvals[i], keyvals[i])
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		it := cache.Iterator(nil, nil)
		for it.Valid() {
			it.Next()
		}
	}
}

func testIterate(low, high []byte, reverse bool, keys ...string) []string {
	kvc := NewKVCache()
	for _, k := range keys {
		bs := []byte(k)
		kvc.Set(bs, bs)
	}
	var it KVIterator
	if reverse {
		it = kvc.ReverseIterator(low, high)
	} else {
		it = kvc.Iterator(low, high)
	}
	keysOut := []string{}
	for it.Valid() {
		keysOut = append(keysOut, string(it.Value()))
		it.Next()
	}
	return keysOut
}
