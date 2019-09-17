package storage

import (
	bin "encoding/binary"
	"math/rand"
	"testing"

	"github.com/hyperledger/burrow/storage"

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
	assert.Equal(t, []string{"b"}, testIterate([]byte("b"), []byte("c"), false, "a", "b", "c", "d"))
	assert.Equal(t, []string{"b", "c"}, testIterate([]byte("b"), []byte("cc"), false, "a", "b", "c", "d"))
	assert.Equal(t, []string{"a", "b", "c", "d"}, testIterate([]byte(""), nil, false, "a", "b", "c", "d"))
	assert.Equal(t, []string{"d", "c", "b", "a"}, testIterate([]byte(""), nil, true, "a", "b", "c", "d"))
	assert.Equal(t, []string{"a", "b", "c", "d"}, testIterate(nil, nil, false, "a", "b", "c", "d"))

	assert.Equal(t, []string{}, testIterate([]byte(""), []byte(""), false, "a", "b", "c", "d"))
	assert.Equal(t, []string{}, testIterate([]byte("ab"), []byte("ab"), false, "a", "b", "c", "d"))
	assert.Equal(t, []string{"a"}, testIterate([]byte("0"), []byte("ab"), true, "a", "b", "c", "d"))
	assert.Equal(t, []string{"c", "b", "a"}, testIterate([]byte("a"), []byte("c1"), true, "a", "b", "c", "d"))
	assert.Equal(t, []string{"b", "a"}, testIterate([]byte("a"), []byte("c"), true, "a", "b", "c", "d"))
	assert.Equal(t, []string{"b", "a"}, testIterate([]byte("a"), []byte("c"), true, "a", "b", "c", "d"))
	assert.Equal(t, []string{}, testIterate([]byte("c"), []byte("e"), true, "a", "b"))
	assert.Equal(t, []string{}, testIterate([]byte("c"), []byte("e"), true, "z", "f"))
}

func BenchmarkKVCache_Iterator_1E6_Inserts(b *testing.B) {
	benchmarkKVCache_Iterator(b, 1e6)
}

func BenchmarkKVCache_Iterator_1E7_Inserts(b *testing.B) {
	benchmarkKVCache_Iterator(b, 1e7)
}

func benchmarkKVCache_Iterator(b *testing.B, inserts int) {
	b.StopTimer()
	cache := NewKVCache()
	rnd := rand.NewSource(23425)
	keyvals := make([][]byte, inserts)
	for i := 0; i < inserts; i++ {
		bs := make([]byte, 8)
		bin.BigEndian.PutUint64(bs, uint64(rnd.Int63()))
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
	var it storage.KVIterator
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
