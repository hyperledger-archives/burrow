package web3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHexEncoder_BytesTrim(t *testing.T) {
	assert.Equal(t, "", HexEncoder.BytesTrim(nil))
	assert.Equal(t, "", HexEncoder.BytesTrim([]byte{}))
	assert.Equal(t, "0x0", HexEncoder.BytesTrim([]byte{0}))
	assert.Equal(t, "0x1", HexEncoder.BytesTrim([]byte{1}))
	assert.Equal(t, "0x1ff", HexEncoder.BytesTrim([]byte{1, 255}))
}

func TestHexDecoder_Bytes(t *testing.T) {
	d := new(HexDecoder)
	assert.Equal(t, []byte{}, d.Bytes(""))
	assert.Equal(t, []byte{1}, d.Bytes("0x1"))
	assert.Equal(t, []byte{1}, d.Bytes("0x01"))
	assert.Equal(t, []byte{1, 0xff}, d.Bytes("0x1ff"))
	require.NoError(t, d.Err())
}
