package storage

import (
	"bytes"
)

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
