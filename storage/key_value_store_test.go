package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareKeys(t *testing.T) {
	assert.Equal(t, 1, CompareKeys(nil, []byte{2}))
	assert.Equal(t, -1, CompareKeys([]byte{2}, nil))
	assert.Equal(t, -1, CompareKeys([]byte{}, nil))
	assert.Equal(t, 1, CompareKeys(nil, []byte{}))
	assert.Equal(t, 0, CompareKeys(nil, nil))
	assert.Equal(t, -1, CompareKeys([]byte{1, 2, 3}, []byte{2}))
}
