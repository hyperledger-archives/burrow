package evm

import (
	"testing"

	"github.com/hyperledger/burrow/execution/errors"
	"github.com/stretchr/testify/require"

	"math/big"

	"github.com/stretchr/testify/assert"
)

// Test static memory allocation with maximum == initial capacity - memory should not grow
func TestDynamicMemory_StaticAllocation(t *testing.T) {
	err := errors.FirstOnly()
	mem := NewDynamicMemory(4, 4, err).(*dynamicMemory)
	mem.Write(big.NewInt(0), []byte{1})
	mem.Write(big.NewInt(1), []byte{0, 0, 1})
	assert.Equal(t, []byte{1, 0, 0, 1}, mem.slice)
	assert.Equal(t, 4, cap(mem.slice), "Slice capacity should not grow")
	require.NoError(t, err.Error())
}

// Test reading beyond the current capacity - memory should grow
func TestDynamicMemory_ReadAhead(t *testing.T) {
	err := errors.FirstOnly()
	mem := NewDynamicMemory(4, 8, err).(*dynamicMemory)
	value := mem.Read(big.NewInt(2), big.NewInt(4))
	require.NoError(t, err.Error())
	// Value should be size requested
	assert.Equal(t, []byte{0, 0, 0, 0}, value)
	// Slice should have grown to that plus offset
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0}, mem.slice)

	err.Reset()
	value = mem.Read(big.NewInt(2), big.NewInt(6))
	require.NoError(t, err.Error())
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0}, value)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0}, mem.slice)

	// Check cannot read out of bounds
	mem.Read(big.NewInt(2), big.NewInt(7))
	assert.Error(t, err.Error())
}

// Test writing beyond the current capacity - memory should grow
func TestDynamicMemory_WriteAhead(t *testing.T) {
	err := errors.FirstOnly()
	mem := NewDynamicMemory(4, 8, err).(*dynamicMemory)
	mem.Write(big.NewInt(4), []byte{1, 2, 3, 4})
	require.NoError(t, err.Error())
	assert.Equal(t, []byte{0, 0, 0, 0, 1, 2, 3, 4}, mem.slice)

	mem.Write(big.NewInt(4), []byte{1, 2, 3, 4, 5})
	assert.Error(t, err.Error())
}

func TestDynamicMemory_WriteRead(t *testing.T) {
	err := errors.FirstOnly()
	mem := NewDynamicMemory(1, 0x10000000, err).(*dynamicMemory)
	// Text is out of copyright
	bytesToWrite := []byte(`He paused. He felt the rhythm of the verse about him in the room.
How melancholy it was! Could he, too, write like that, express the
melancholy of his soul in verse? There were so many things he wanted
to describe: his sensation of a few hours before on Grattan Bridge, for
example. If he could get back again into that mood....`)

	// Write the bytes
	offset := big.NewInt(0x1000000)
	mem.Write(offset, bytesToWrite)
	require.NoError(t, err.Error())
	assert.Equal(t, append(make([]byte, offset.Uint64()), bytesToWrite...), mem.slice)
	assert.Equal(t, offset.Uint64()+uint64(len(bytesToWrite)), uint64(len(mem.slice)))

	// Read them back
	value := mem.Read(offset, big.NewInt(int64(len(bytesToWrite))))
	require.NoError(t, err.Error())
	assert.Equal(t, bytesToWrite, value)
}

func TestDynamicMemory_ZeroInitialMemory(t *testing.T) {
	err := errors.FirstOnly()
	mem := NewDynamicMemory(0, 16, err).(*dynamicMemory)
	mem.Write(big.NewInt(4), []byte{1, 2, 3, 4})
	require.NoError(t, err.Error())
	assert.Equal(t, []byte{0, 0, 0, 0, 1, 2, 3, 4}, mem.slice)
}

func TestDynamicMemory_Capacity(t *testing.T) {
	err := errors.FirstOnly()
	mem := NewDynamicMemory(1, 0x10000000, err).(*dynamicMemory)

	assert.Equal(t, big.NewInt(1), mem.Capacity())

	capacity := big.NewInt(1234)
	mem.ensureCapacity(capacity.Uint64())
	require.NoError(t, err.Error())
	assert.Equal(t, capacity, mem.Capacity())

	capacity = big.NewInt(123456789)
	mem.ensureCapacity(capacity.Uint64())
	require.NoError(t, err.Error())
	assert.Equal(t, capacity, mem.Capacity())

	// Check doesn't shrink or err
	mem.ensureCapacity(12)
	require.NoError(t, err.Error())
	assert.Equal(t, capacity, mem.Capacity())
}

func TestDynamicMemory_ensureCapacity(t *testing.T) {
	mem := NewDynamicMemory(4, 16, errors.FirstOnly()).(*dynamicMemory)
	// Check we can grow within bounds
	err := mem.ensureCapacity(8)
	require.NoError(t, err)
	expected := make([]byte, 8)
	assert.Equal(t, expected, mem.slice)

	// Check we can grow to bounds
	err = mem.ensureCapacity(16)
	require.NoError(t, err)
	expected = make([]byte, 16)
	assert.Equal(t, expected, mem.slice)

	err = mem.ensureCapacity(1)
	require.NoError(t, err)
	assert.Equal(t, 16, len(mem.slice))

	err = mem.ensureCapacity(17)
	assert.Error(t, err, "Should not be possible to grow over capacity")

}
