package storage

import "testing"

func TestNewChannelIterator(t *testing.T) {
	ch := make(chan KVPair)
	go sendKVPair(ch, kvPairs("a", "hello", "b", "channel", "c", "this is nice"))
	ci := NewChannelIterator(ch, bz("a"), bz("c"))
	checkItem(t, ci, bz("a"), bz("hello"))
	checkNext(t, ci, true)
	checkItem(t, ci, bz("b"), bz("channel"))
	checkNext(t, ci, true)
	checkItem(t, ci, bz("c"), bz("this is nice"))
	checkNext(t, ci, false)
	checkInvalid(t, ci)
}
