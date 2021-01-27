package storage

import (
	dbm "github.com/tendermint/tm-db"
)

type KVIterator = dbm.Iterator

// This is partially extracted from Cosmos SDK for alignment but is more minimal, we should suggest this becomes an
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

type Versioned interface {
	Hash() []byte
	Version() int64
}
