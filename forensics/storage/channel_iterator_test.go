package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tm-db"
)

func TestNewChannelIterator(t *testing.T) {
	ch := make(chan KVPair)
	go sendKVPair(ch, kvPairs("a", "hello", "b", "channel", "c", "this is nice"))
	ci := NewChannelIterator(ch, []byte("a"), []byte("c"))
	checkItem(t, ci, []byte("a"), []byte("hello"))
	checkNext(t, ci, true)
	checkItem(t, ci, []byte("b"), []byte("channel"))
	checkNext(t, ci, true)
	checkItem(t, ci, []byte("c"), []byte("this is nice"))
	checkNext(t, ci, false)
	checkInvalid(t, ci)
}

func checkInvalid(t *testing.T, itr dbm.Iterator) {
	checkValid(t, itr, false)
	checkKeyPanics(t, itr)
	checkValuePanics(t, itr)
	checkNextPanics(t, itr)
}

func checkValid(t *testing.T, itr dbm.Iterator, expected bool) {
	valid := itr.Valid()
	require.Equal(t, expected, valid)
}

func checkNext(t *testing.T, itr dbm.Iterator, expected bool) {
	itr.Next()
	valid := itr.Valid()
	require.Equal(t, expected, valid)
}

func checkNextPanics(t *testing.T, itr dbm.Iterator) {
	assert.Panics(t, func() { itr.Next() }, "checkNextPanics expected panic but didn't")
}
func checkKeyPanics(t *testing.T, itr dbm.Iterator) {
	assert.Panics(t, func() { itr.Key() }, "checkKeyPanics expected panic but didn't")
}

func checkValuePanics(t *testing.T, itr dbm.Iterator) {
	assert.Panics(t, func() { itr.Key() }, "checkValuePanics expected panic but didn't")
}
