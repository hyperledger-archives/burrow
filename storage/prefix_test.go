package storage

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestPrefix_Iterable(t *testing.T) {
	keys := [][]byte{
		{0x10, 0xab},
		{0x11, 0x00},
		{0x11, 0x00, 0x00},
		{0x11, 0x00, 0x00, 1},
		{0x11, 0x00, 0x00, 2},
		{0x11, 0x00, 0x00, 3},
		{0x11, 0x00, 0x00, 4},
		{0x11, 0x00, 0x01, 0x00},
		{0x11, 0x34, 0x00},
		{0x11, 0xff, 0xff},
		{0x11, 0xff, 0xff, 0xff},
		{0x12},
	}
	memDB := dbm.NewMemDB()
	for i, k := range keys {
		memDB.Set(k, []byte{byte(i)})
	}
	requireKeysSorted(t, keys)
	p := Prefix([]byte{0x11, 0x00, 0x00})
	it := p.Iterable(memDB)
	expectedKeys := [][]byte{{}, {1}, {2}, {3}, {4}}
	requireKeysSorted(t, expectedKeys)
	assert.Equal(t, expectedKeys, dumpKeys(it.Iterator(nil, nil)))

	expectedKeys = [][]byte{{4}, {3}, {2}, {1}, {}}
	requireKeysSorted(t, expectedKeys, true)
	assert.Equal(t, expectedKeys, dumpKeys(it.ReverseIterator(nil, nil)))
}

func requireKeysSorted(t *testing.T, keys [][]byte, reverse ...bool) {
	comp := -1
	if len(reverse) > 0 && reverse[0] {
		comp = 1
	}
	sortedKeys := make([][]byte, len(keys))
	for i, k := range keys {
		sortedKeys[i] = make([]byte, len(k))
		copy(sortedKeys[i], k)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return strings.Compare(string(sortedKeys[i]), string(sortedKeys[j])) == comp
	})
	require.Equal(t, sortedKeys, keys)
}

func dumpKeys(it dbm.Iterator) [][]byte {
	var keys [][]byte
	for it.Valid() {
		keys = append(keys, it.Key())
		it.Next()
	}
	return keys
}
