package rpcevents

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockRange_Bounds(t *testing.T) {
	latestHeight := uint64(2344)
	br := &BlockRange{}
	start, end, streaming := br.Bounds(latestHeight)
	assert.Equal(t, latestHeight, start)
	assert.Equal(t, latestHeight, end)
	assert.False(t, streaming)
}
