package exec

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
)

func TestBlockExecution_Marshal(t *testing.T) {
	be := &BlockExecution{
		Header: &types.Header{
			Height:          3,
			AppHash:         []byte{2},
			ProposerAddress: []byte{1, 2, 33},
		},
	}
	bs, err := be.Marshal()
	require.NoError(t, err)
	beOut := new(BlockExecution)
	require.NoError(t, beOut.Unmarshal(bs))
}
