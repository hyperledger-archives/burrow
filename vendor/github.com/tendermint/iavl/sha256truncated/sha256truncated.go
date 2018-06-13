// Package sha256truncated provides a sha256 hash.Hash whose output is truncated to 20 bytes (160 bits).
//
// This is the default hashing algorithm used by IAVL+ trees.
//
//   s256 := sha256.New() // crypto/sha256
//   s256Truncated := New() // this package
//
//   // Use like any other hash.Hash ...
//   // Contract:
//   s256Trunc.Sum(nil) == s256.Sum(nil)[:20]
package sha256truncated

import (
	"crypto/sha256"
	"hash"
)

const Size = 20

// New returns a new hash.Hash computing the truncated to the first 20 bytes SHA256 checksum.
func New() hash.Hash {
	return &digest{sha256.New()}
}

func (d *digest) Sum(in []byte) []byte {
	return d.Hash.Sum(in)[:Size]
}

func (d *digest) Reset() {
	d.Hash.Reset()
}

func (d *digest) Size() int {
	return Size
}

func (d *digest) BlockSize() int {
	return d.Hash.BlockSize()
}

// digest is just a wrapper around sha256
type digest struct {
	hash.Hash
}
