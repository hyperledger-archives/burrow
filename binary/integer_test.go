package binary

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUint64SumOverflow(t *testing.T) {
	var b uint64 = 0xdeadbeef
	var a uint64 = math.MaxUint64 - b
	assert.False(t, IsUint64SumOverflow(a-b, b))
	assert.False(t, IsUint64SumOverflow(a, b))
	assert.False(t, IsUint64SumOverflow(a+b, 0))
	assert.True(t, IsUint64SumOverflow(a, b+1))
	assert.True(t, IsUint64SumOverflow(a+b, 1))
	assert.True(t, IsUint64SumOverflow(a+1, b+1))
}
