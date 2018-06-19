package events

import (
	"bytes"
	"fmt"

	"github.com/hyperledger/burrow/binary"
)

type Key []byte

func NewKey(height, index uint64) Key {
	k := make(Key, 16)
	// Will order first by height then by index so events from the same block will be consecutive
	binary.PutUint64BE(k[:8], height)
	binary.PutUint64BE(k[8:], index)
	return k
}

// -1 if k < k2
//  0 if k == k2
//  1 if k > k2
func (k Key) Compare(k2 Key) int {
	return bytes.Compare(k, k2)
}

// Returns true iff k is a valid successor key to p;
// iff (the height is the same and the index is one greater) or (the height is greater and the index is zero) or (p
// is uninitialised)
func (k Key) IsSuccessorOf(p Key) bool {
	if len(p) == 0 {
		return true
	}
	ph, kh := p.Height(), k.Height()
	pi, ki := p.Index(), k.Index()
	return ph == kh && pi+1 == ki || ph < kh && ki == 0
}

func (k Key) Bytes() []byte {
	return k
}

func (k Key) Height() uint64 {
	return binary.GetUint64BE(k[:8])
}

func (k Key) Index() uint64 {
	return binary.GetUint64BE(k[8:])
}

func (k Key) String() string {
	return fmt.Sprintf("Key{Height: %v; Index: %v}", k.Height(), k.Index())
}
