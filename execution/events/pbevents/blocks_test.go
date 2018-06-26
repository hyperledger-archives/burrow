package pbevents

import (
	"testing"

	"github.com/hyperledger/burrow/execution/events"
	"github.com/stretchr/testify/assert"
)

func TestBlockRange_Bounds(t *testing.T) {
	latestHeight := uint64(2344)
	br := &BlockRange{}
	start, end, streaming := br.Bounds(latestHeight)
	assert.Equal(t, events.NewKey(latestHeight, 0), start)
	assert.Equal(t, events.NewKey(latestHeight, 0), end)
	assert.False(t, streaming)
}
