package sync

import (
	"sync"

	"hash"

	"encoding/binary"

	"github.com/OneOfOne/xxhash"
)

type RingMutex struct {
	mtxs       []sync.RWMutex
	hash       func(address []byte) uint64
	mutexCount uint64
}

// Create a RW mutex that provides a pseudo-independent set of mutexes for addresses
// where the address space is mapped into possibly much smaller set of backing
// mutexes using the xxhash (non-cryptographic)
// hash function // modulo size. If some addresses collide modulo size they will be unnecessary
// contention between those addresses, but you can trade space against contention
// as desired.
func NewRingMutex(mutexCount int, hashMaker func() hash.Hash64) *RingMutex {
	ringMutex := &RingMutex{
		mutexCount: uint64(mutexCount),
		// max slice length is bounded by max(int) thus the argument type
		mtxs: make([]sync.RWMutex, mutexCount, mutexCount),
		hash: func(address []byte) uint64 {
			buf := make([]byte, 8)
			copy(buf, address)
			return binary.LittleEndian.Uint64(buf)
		},
	}
	if hashMaker != nil {
		hasherPool := &sync.Pool{
			New: func() interface{} {
				return hashMaker()
			},
		}
		ringMutex.hash = func(address []byte) uint64 {
			h := hasherPool.Get().(hash.Hash64)
			defer func() {
				h.Reset()
				hasherPool.Put(h)
			}()
			h.Write(address)
			return h.Sum64()
		}
	}
	return ringMutex
}

func NewRingMutexNoHash(mutexCount int) *RingMutex {
	return NewRingMutex(mutexCount, nil)
}

func NewRingMutexXXHash(mutexCount int) *RingMutex {
	return NewRingMutex(mutexCount, func() hash.Hash64 {
		return xxhash.New64()
	})
}

func (mtx *RingMutex) Lock(address []byte) {
	mtx.Mutex(address).Lock()
}

func (mtx *RingMutex) Unlock(address []byte) {
	mtx.Mutex(address).Unlock()
}

func (mtx *RingMutex) RLock(address []byte) {
	mtx.Mutex(address).RLock()
}

func (mtx *RingMutex) RUnlock(address []byte) {
	mtx.Mutex(address).RUnlock()
}

// Return the size of the underlying array of mutexes
func (mtx *RingMutex) MutexCount() uint64 {
	return mtx.mutexCount
}

func (mtx *RingMutex) Mutex(address []byte) *sync.RWMutex {
	return &mtx.mtxs[mtx.index(address)]
}

func (mtx *RingMutex) index(address []byte) uint64 {
	return mtx.hash(address) % mtx.mutexCount
}
