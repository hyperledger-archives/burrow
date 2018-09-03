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
