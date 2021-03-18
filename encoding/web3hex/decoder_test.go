package web3hex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecoder_Bytes(t *testing.T) {
	d := new(Decoder)
	assert.Equal(t, []byte{}, d.Bytes(""))
	assert.Equal(t, []byte{1}, d.Bytes("0x1"))
	assert.Equal(t, []byte{1}, d.Bytes("0x01"))
	assert.Equal(t, []byte{1, 0xff}, d.Bytes("0x1ff"))
	require.NoError(t, d.Err())
}
