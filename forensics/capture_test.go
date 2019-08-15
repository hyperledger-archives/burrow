package forensics

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/binary"
)

func TestCompareCapture(t *testing.T) {
	exp := []*ReplayCapture{{
		Height:        0,
		AppHashBefore: binary.HexBytes("00000000000000000000"),
		AppHashAfter:  binary.HexBytes("00000000000000000000"),
	}, {
		Height:        1,
		AppHashBefore: binary.HexBytes("00000000000000000000"),
		AppHashAfter:  binary.HexBytes("00000000000000000000"),
	}}
	act := []*ReplayCapture{{
		Height:        0,
		AppHashBefore: binary.HexBytes("00000000000000000000"),
		AppHashAfter:  binary.HexBytes("00000000000000000000"),
	}, {
		Height:        1,
		AppHashBefore: binary.HexBytes("00000000000000000000"),
		AppHashAfter:  binary.HexBytes("11111111111111111111"),
	}}
	height, err := CompareCaptures(exp, act)
	require.Error(t, err)
	require.Equal(t, uint64(1), height)
}
