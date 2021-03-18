package encoding

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEthChainID(t *testing.T) {
	assert.Equal(t, big.NewInt(1234), GetEthChainID("1234"))
	assert.Equal(t, big.NewInt(1234), GetEthChainID("0x4d2"))
	chainID, ok := new(big.Int).SetString("28980219985052679991929851741845949978287371722649499714751652210", 10)
	require.True(t, ok)
	assert.Equal(t, chainID, GetEthChainID("FrogsEatApplesOnlyWhenClear"))
}
