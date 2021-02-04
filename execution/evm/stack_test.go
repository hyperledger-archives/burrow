package evm

import (
	"math"
	"math/big"
	"testing"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var maxUint64 = big.NewInt(math.MaxInt64)

func TestStack_MaxDepthInt32(t *testing.T) {
	st := NewStack(new(errors.Maybe), 0, 0, maxUint64)

	err := st.ensureCapacity(math.MaxInt32 + 1)
	assert.Error(t, err)
}

// Test static memory allocation with unlimited depth - memory should grow
func TestStack_UnlimitedAllocation(t *testing.T) {
	err := new(errors.Maybe)
	st := NewStack(err, 0, 0, maxUint64)

	st.Push64(math.MaxInt64)
	require.NoError(t, err.Error())
	assert.Equal(t, 1, len(st.slice))
	assert.Equal(t, 1, cap(st.slice))
}

// Test static memory allocation with maximum == initial capacity - memory should not grow
func TestStack_StaticAllocation(t *testing.T) {
	err := new(errors.Maybe)

	st := NewStack(err, 4, 4, maxUint64)

	for i := 0; i < 4; i++ {
		st.Push64(math.MaxInt64)
		assert.NoError(t, err.Error())
	}

	assert.Equal(t, 4, cap(st.slice), "Slice capacity should not grow")
}

// Test writing beyond the current capacity - memory should grow
func TestDynamicMemory_PushAhead(t *testing.T) {
	err := new(errors.Maybe)

	st := NewStack(err, 2, 4, maxUint64)

	for i := 0; i < 4; i++ {
		st.Push64(math.MaxInt64)
		assert.NoError(t, err.Error())
	}

	st.Push64(math.MaxInt64)
	assert.Equal(t, errors.Codes.DataStackOverflow, errors.GetCode(err.Error()))
}

func TestStack_ZeroInitialCapacity(t *testing.T) {
	err := new(errors.Maybe)

	st := NewStack(err, 0, 16, maxUint64)
	require.NoError(t, err.Error())
	st.Push64(math.MaxInt64)
	assert.Equal(t, []binary.Word256{binary.Int64ToWord256(math.MaxInt64)}, st.slice)
}

func TestStack_ensureCapacity(t *testing.T) {

	st := NewStack(new(errors.Maybe), 4, 16, maxUint64)
	// Check we can grow within bounds
	err := st.ensureCapacity(8)
	assert.NoError(t, err)
	expected := make([]binary.Word256, 8)
	assert.Equal(t, expected, st.slice)

	// Check we can grow to bounds
	err = st.ensureCapacity(16)
	assert.NoError(t, err)
	expected = make([]binary.Word256, 16)
	assert.Equal(t, expected, st.slice)

	err = st.ensureCapacity(1)
	assert.NoError(t, err)
	assert.Equal(t, 16, len(st.slice))

	err = st.ensureCapacity(17)
	assert.Error(t, err, "Should not be possible to grow over capacity")
}
