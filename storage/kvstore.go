package storage

import (
	"bytes"

	dbm "github.com/tendermint/tm-db"
)

type KVIterator = dbm.Iterator

// This is partially extrated from Cosmos SDK for alignment but is more minimal, we should suggest this becomes an
// embedded interface
type KVIterable interface {
	// Iterator over a domain of keys in ascending order. high is exclusive.
	// low must be less than high, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	Iterator(low, high []byte) (KVIterator, error)

	// Iterator over a domain of keys in descending order. high is exclusive.
	//  must be less than high, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	ReverseIterator(low, high []byte) (KVIterator, error)
}

// Provides the native iteration for IAVLTree
type KVCallbackIterable interface {
	// low must be lexicographically less than high. high is exclusive unless it is nil in which case it is inclusive.
	// ascending == false reverses order.
	Iterate(low, high []byte, ascending bool, fn func(key []byte, value []byte) error) error
}

type KVReader interface {
	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) ([]byte, error)
	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) (bool, error)
}

type KVWriter interface {
	// Set sets the key. Panics on nil key.
	Set(key, value []byte) error
	// Delete deletes the key. Panics on nil key.
	Delete(key []byte) error
}

type KVIterableReader interface {
	KVReader
	KVIterable
}

type KVCallbackIterableReader interface {
	KVReader
	KVCallbackIterable
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

// NormaliseDomain encodes the assumption that when nil is used as a lower bound is interpreted as low rather than high
func NormaliseDomain(low, high []byte) ([]byte, []byte) {
	if len(low) == 0 {
		low = []byte{}
	}
	return low, high
}
