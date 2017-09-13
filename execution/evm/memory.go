package vm

import (
	"fmt"
	"math"
)

const (
	defaultInitialMemoryCapacity = 0x100000  // 1 MiB
	defaultMaximumMemoryCapacity = 0x1000000 // 16 MiB
)

// Change the length of this zero array to tweak the size of the block of zeros
// written to the backing slice at a time when it is grown. A larger number may
// lead to fewer calls to append to achieve the desired capacity although it is
// unlikely to make a lot of difference.
var zeroBlock []byte = make([]byte, 32)

// Interface for a bounded linear memory indexed by a single int64 parameter
// for each byte in the memory.
type Memory interface {
	// Read a value from the memory store starting at offset
	// (index of first byte will equal offset). The value will be returned as a
	// length-bytes byte slice. Returns an error if the memory cannot be read or
	// is not allocated.
	//
	// The value returned should be copy of any underlying memory, not a reference
	// to the underlying store.
	Read(offset, length int64) ([]byte, error)
	// Write a value to the memory starting at offset (the index of the first byte
	// written will equal offset). The value is provided as bytes to be written
	// consecutively to the memory store. Return an error if the memory cannot be
	// written or allocated.
	Write(offset int64, value []byte) error
	// Returns the current capacity of the memory. For dynamically allocating
	// memory this capacity can be used as a write offset that is guaranteed to be
	// unused. Solidity in particular makes this assumption when using MSIZE to
	// get the current allocated memory.
	Capacity() int64
}

func NewDynamicMemory(initialCapacity, maximumCapacity int64) Memory {
	return &dynamicMemory{
		slice:           make([]byte, initialCapacity),
		maximumCapacity: maximumCapacity,
	}
}

func DefaultDynamicMemoryProvider() Memory {
	return NewDynamicMemory(defaultInitialMemoryCapacity, defaultMaximumMemoryCapacity)
}

// Implements a bounded dynamic memory that relies on Go's (pretty good) dynamic
// array allocation via a backing slice
type dynamicMemory struct {
	slice           []byte
	maximumCapacity int64
}

func (mem *dynamicMemory) Read(offset, length int64) ([]byte, error) {
	capacity := offset + length
	err := mem.ensureCapacity(capacity)
	if err != nil {
		return nil, err
	}
	value := make([]byte, length)
	copy(value, mem.slice[offset:capacity])
	return value, nil
}

func (mem *dynamicMemory) Write(offset int64, value []byte) error {
	capacity := offset + int64(len(value))
	err := mem.ensureCapacity(capacity)
	if err != nil {
		return err
	}
	copy(mem.slice[offset:capacity], value)
	return nil
}

func (mem *dynamicMemory) Capacity() int64 {
	return int64(len(mem.slice))
}

// Ensures the current memory store can hold newCapacity. Will only grow the
// memory (will not shrink).
func (mem *dynamicMemory) ensureCapacity(newCapacity int64) error {
	if newCapacity > math.MaxInt32 {
		// If we ever did want more than an int32 of space then we would need to
		// maintain multiple pages of memory
		return fmt.Errorf("cannot address memory beyond a maximum index "+
			"of Int32 type (%v bytes)", math.MaxInt32)
	}
	newCapacityInt := int(newCapacity)
	// We're already big enough so return
	if newCapacityInt <= len(mem.slice) {
		return nil
	}
	if newCapacity > mem.maximumCapacity {
		return fmt.Errorf("cannot grow memory because it would exceed the "+
			"current maximum limit of %v bytes", mem.maximumCapacity)
	}
	// Ensure the backing array of slice is big enough
	// Grow the memory one word at time using the pre-allocated zeroBlock to avoid
	// unnecessary allocations. Use append to make use of any spare capacity in
	// the slice's backing array.
	for newCapacityInt > cap(mem.slice) {
		// We'll trust Go exponentially grow our arrays (at first).
		mem.slice = append(mem.slice, zeroBlock...)
	}
	// Now we've ensured the backing array of the slice is big enough we can
	// just re-slice (even if len(mem.slice) < newCapacity)
	mem.slice = mem.slice[:newCapacity]
	return nil
}
