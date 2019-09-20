package encoding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeNumber(t *testing.T) {
	require.Equal(t, "0x0", EncodeNumber(0))
	i, err := DecodeToNumber("0x0")
	require.NoError(t, err)
	require.Equal(t, uint64(0), i)

	require.Equal(t, "0x5208", EncodeNumber(21000))
	i, err = DecodeToNumber("0x5208")
	require.NoError(t, err)
	require.Equal(t, uint64(21000), i)
}

func TestEncodeBytes(t *testing.T) {
	require.Equal(t, "0x68656c6c6f2c20776f726c64", EncodeBytes([]byte("hello, world")))
	b, err := DecodeToBytes("0x68656c6c6f2c20776f726c64")
	require.NoError(t, err)
	require.Equal(t, []byte("hello, world"), b)
}
