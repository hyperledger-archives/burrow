package abi

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEVMInt(t *testing.T) {
	t.Run("pack big.Int", func(t *testing.T) {
		e := EVMInt{256}
		b := big.NewInt(-23423)
		data, err := e.pack(b)
		require.NoError(t, err)
		bOut := new(big.Int)
		_, err = e.unpack(data, 0, &bOut)
		require.NoError(t, err)
		assert.Equal(t, bOut, b)
	})
}
