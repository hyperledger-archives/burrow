package storage

import (
	"bytes"

	dbm "github.com/tendermint/tendermint/libs/db"
)

type KVIterator = dbm.Iterator

// This is partially extrated from Cosmos SDK for alignment but is more minimal, we should suggest this becomes an
// embedded interface
type KVIterable interface {
	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	Iterator(start, end []byte) KVIterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	ReverseIterator(start, end []byte) KVIterator
}

type KVReader interface {
	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte
	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool
}

type KVWriter interface {
	// Set sets the key. Panics on nil key.
	Set(key, value []byte)
	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)
}

type KVIterableReader interface {
	KVReader
	KVIterable
}

// KVStore is a simple interface to get/set data
type KVReaderWriter interface {
	KVReader
	KVWriter
}

type KVStore interface {
	KVReaderWriter
	KVIterable
}

// NormaliseDomain encodes the assumption that when nil is used as a lower bound is interpreted as low rather
// than high
func NormaliseDomain(start, end []byte, reverse bool) ([]byte, []byte) {
	if reverse {
		if len(end) == 0 {
			end = []byte{}
		}
	} else {
		if len(start) == 0 {
			start = []byte{}
		}
	}
	return start, end
}

// KeyOrder maps []byte{} -> -1, []byte(nil) -> 1, and everything else to 0. This encodes the assumptions of the
// KVIterator domain endpoints
func KeyOrder(key []byte) int {
	if key == nil {
		// Sup
		return 1
	}
	if len(key) == 0 {
		// Inf
		return -1
	}
	// Normal key
	return 0
}

// Sorts the keys as if they were compared lexicographically with their KeyOrder prepended
func CompareKeys(k1, k2 []byte) int {
	ko1 := KeyOrder(k1)
	ko2 := KeyOrder(k2)
	if ko1 < ko2 {
		return -1
	}
	if ko1 > ko2 {
		return 1
	}
	return bytes.Compare(k1, k2)
}
