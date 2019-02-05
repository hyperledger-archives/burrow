package storage

import (
	"math/rand"
	"testing"

	"github.com/hyperledger/burrow/binary"

	"github.com/stretchr/testify/assert"
)

func TestKVCache_Iterator(t *testing.T) {
	kvc := NewKVCache()
	kvp := kvPairs("b", "ar", "f", "oo", "im", "aginative")
	for _, kv := range kvp {
		kvc.Set(kv.Key, kv.Value)
	}
	assert.Equal(t, kvp, collectIterator(kvc.Iterator(nil, nil)))
}

func TestKVCache_SortedKeysInDomain(t *testing.T) {
	assert.Equal(t, []string{"b"}, testSortedKeysInDomain(bz("b"), bz("c"), "a", "b", "c", "d"))
	assert.Equal(t, []string{"b", "c"}, testSortedKeysInDomain(bz("b"), bz("cc"), "a", "b", "c", "d"))
	assert.Equal(t, []string{"a", "b", "c", "d"}, testSortedKeysInDomain(bz(""), nil, "a", "b", "c", "d"))
	assert.Equal(t, []string{"d", "c", "b", "a"}, testSortedKeysInDomain(nil, bz(""), "a", "b", "c", "d"))
	assert.Equal(t, []string{}, testSortedKeysInDomain(nil, nil, "a", "b", "c", "d"))
	assert.Equal(t, []string{}, testSortedKeysInDomain(bz(""), bz(""), "a", "b", "c", "d"))
	assert.Equal(t, []string{}, testSortedKeysInDomain(bz("ab"), bz("ab"), "a", "b", "c", "d"))
	assert.Equal(t, []string{"a"}, testSortedKeysInDomain(bz("0"), bz("ab"), "a", "b", "c", "d"))
	assert.Equal(t, []string{"c", "b"}, testSortedKeysInDomain(bz("c1"), bz("a"), "a", "b", "c", "d"))
	assert.Equal(t, []string{"c", "b"}, testSortedKeysInDomain(bz("c"), bz("a"), "a", "b", "c", "d"))
	assert.Equal(t, []string{}, testSortedKeysInDomain(bz("e"), bz("c"), "a", "b"))
	assert.Equal(t, []string{}, testSortedKeysInDomain(bz("e"), bz("c"), "z", "f"))
}

func BenchmarkKVCache_Iterator(b *testing.B) {
	b.StopTimer()
	cache := NewKVCache()
	rnd := rand.NewSource(23425)
	keyvals := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		bs := make([]byte, 8)
		binary.PutInt64BE(bs, rnd.Int63())
		keyvals[i] = bs
	}
	for i := 0; i < b.N; i++ {
		cache.Set(keyvals[i], keyvals[i])
	}
	b.StartTimer()
	it := cache.Iterator(nil, nil)
	for it.Valid() {
		it.Next()
	}
}

func testSortedKeysInDomain(start, end []byte, keys ...string) []string {
	cache := make(map[string]valueInfo)
	for _, k := range keys {
		cache[k] = valueInfo{}
	}
	kvc := KVCache{
		cache: cache,
	}
	bkeys := kvc.SortedKeysInDomain(start, end)
	keys = make([]string, len(bkeys))
	for i, bk := range bkeys {
		keys[i] = string(bk)
	}
	return keys
}
