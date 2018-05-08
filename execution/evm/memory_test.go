package evm

import (
	"testing"

	"math/big"

	"github.com/stretchr/testify/assert"
)

// Test static memory allocation with maximum == initial capacity - memory should not grow
func TestDynamicMemory_StaticAllocation(t *testing.T) {
	mem := NewDynamicMemory(4, 4).(*dynamicMemory)
	mem.Write(big.NewInt(0), []byte{1})
	mem.Write(big.NewInt(1), []byte{0, 0, 1})
	assert.Equal(t, []byte{1, 0, 0, 1}, mem.slice)
	assert.Equal(t, 4, cap(mem.slice), "Slice capacity should not grow")
}

// Test reading beyond the current capacity - memory should grow
func TestDynamicMemory_ReadAhead(t *testing.T) {
	mem := NewDynamicMemory(4, 8).(*dynamicMemory)
	value, err := mem.Read(big.NewInt(2), big.NewInt(4))
	assert.NoError(t, err)
	// Value should be size requested
	assert.Equal(t, []byte{0, 0, 0, 0}, value)
	// Slice should have grown to that plus offset
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0}, mem.slice)

	value, err = mem.Read(big.NewInt(2), big.NewInt(6))
	assert.NoError(t, err)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0}, value)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, mem.slice)

	// Check cannot read out of bounds
	_, err = mem.Read(big.NewInt(2), big.NewInt(7))
	assert.Error(t, err)
}

// Test writing beyond the current capacity - memory should grow
func TestDynamicMemory_WriteAhead(t *testing.T) {
	mem := NewDynamicMemory(4, 8).(*dynamicMemory)
	err := mem.Write(big.NewInt(4), []byte{1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, []byte{0, 0, 0, 0, 1, 2, 3, 4}, mem.slice)

	err = mem.Write(big.NewInt(4), []byte{1, 2, 3, 4, 5})
	assert.Error(t, err)
}

func TestDynamicMemory_WriteRead(t *testing.T) {
	mem := NewDynamicMemory(1, 0x10000000).(*dynamicMemory)
	// Text is out of copyright
	bytesToWrite := []byte(`He paused. He felt the rhythm of the verse about him in the room.
How melancholy it was! Could he, too, write like that, express the
melancholy of his soul in verse? There were so many things he wanted
to describe: his sensation of a few hours before on Grattan Bridge, for
example. If he could get back again into that mood....`)

	// Write the bytes
	offset := big.NewInt(0x1000000)
	err := mem.Write(offset, bytesToWrite)
	assert.NoError(t, err)
	assert.Equal(t, append(make([]byte, offset.Uint64()), bytesToWrite...), mem.slice)
	assert.Equal(t, offset.Uint64()+uint64(len(bytesToWrite)), uint64(len(mem.slice)))

	// Read them back
	value, err := mem.Read(offset, big.NewInt(int64(len(bytesToWrite))))
	assert.NoError(t, err)
	assert.Equal(t, bytesToWrite, value)
}

func TestDynamicMemory_ZeroInitialMemory(t *testing.T) {
	mem := NewDynamicMemory(0, 16).(*dynamicMemory)
	err := mem.Write(big.NewInt(4), []byte{1, 2, 3, 4})
	assert.NoError(t, err)
	assert.Equal(t, []byte{0, 0, 0, 0, 1, 2, 3, 4}, mem.slice)
}

func TestDynamicMemory_Capacity(t *testing.T) {
	mem := NewDynamicMemory(1, 0x10000000).(*dynamicMemory)

	assert.Equal(t, big.NewInt(1), mem.Capacity())

	capacity := big.NewInt(1234)
	err := mem.ensureCapacity(capacity.Uint64())
	assert.NoError(t, err)
	assert.Equal(t, capacity, mem.Capacity())

	capacity = big.NewInt(123456789)
	err = mem.ensureCapacity(capacity.Uint64())
	assert.NoError(t, err)
	assert.Equal(t, capacity, mem.Capacity())

	// Check doesn't shrink or err
	err = mem.ensureCapacity(12)
	assert.NoError(t, err)
	assert.Equal(t, capacity, mem.Capacity())
}

func TestDynamicMemory_ensureCapacity(t *testing.T) {
	mem := NewDynamicMemory(4, 16).(*dynamicMemory)
	// Check we can grow within bounds
	err := mem.ensureCapacity(8)
	assert.NoError(t, err)
	expected := make([]byte, 8)
	assert.Equal(t, expected, mem.slice)

	// Check we can grow to bounds
	err = mem.ensureCapacity(16)
	assert.NoError(t, err)
	expected = make([]byte, 16)
	assert.Equal(t, expected, mem.slice)

	err = mem.ensureCapacity(1)
	assert.NoError(t, err)
	assert.Equal(t, 16, len(mem.slice))

	err = mem.ensureCapacity(17)
	assert.Error(t, err, "Should not be possible to grow over capacity")

}
