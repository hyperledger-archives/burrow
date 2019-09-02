package storage

import (
	"strings"
	"testing"

	"github.com/hyperledger/burrow/storage"
	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tm-db"
)

func sendKVPair(ch chan<- KVPair, kvs []KVPair) {
	for _, kv := range kvs {
		ch <- kv
	}
	close(ch)
}

func collectIterator(it storage.KVIterator) KVPairs {
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

func assertIteratorSorted(t *testing.T, it storage.KVIterator, reverse bool) {
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

func checkItem(t *testing.T, itr dbm.Iterator, key []byte, value []byte) {
	k, v := itr.Key(), itr.Value()
	assert.Exactly(t, key, k)
	assert.Exactly(t, value, v)
}
