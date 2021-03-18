package web3hex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoder_BytesTrim(t *testing.T) {
	assert.Equal(t, "", Encoder.BytesTrim(nil))
	assert.Equal(t, "", Encoder.BytesTrim([]byte{}))
	assert.Equal(t, "0x0", Encoder.BytesTrim([]byte{0}))
	assert.Equal(t, "0x1", Encoder.BytesTrim([]byte{1}))
	assert.Equal(t, "0x1ff", Encoder.BytesTrim([]byte{1, 255}))
}
