package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmthrgd/go-hex"
)

func TestHeader_Key(t *testing.T) {
	h := &Header{
		EventID: "Foos",
		Height:  2345345232,
		Index:   34,
	}
	key := h.Key()
	keyString := hex.EncodeUpperToString(key)
	assert.Equal(t, "000000008BCB20D00000000000000022", keyString)
	assert.Len(t, keyString, 32, "should be 16 bytes")
	assert.Equal(t, h.Height, key.Height())
	assert.Equal(t, h.Index, key.Index())
}

func TestKey_IsSuccessorOf(t *testing.T) {
	assert.True(t, NewKey(1, 0).IsSuccessorOf(NewKey(0, 1)))
	assert.True(t, NewKey(100, 24).IsSuccessorOf(NewKey(100, 23)))
	assert.False(t, NewKey(100, 23).IsSuccessorOf(NewKey(100, 25)))
	assert.False(t, NewKey(1, 1).IsSuccessorOf(NewKey(0, 25)))
	assert.True(t, NewKey(3, 0).IsSuccessorOf(NewKey(2, 0)))
}
