package exec

import (
	"testing"

	"github.com/hyperledger/burrow/event/query"
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

func TestBlockExecution_StreamEvents(t *testing.T) {
	be := &BlockExecution{
		Header: &types.Header{
			Height:          2,
			AppHash:         []byte{2},
			ProposerAddress: []byte{1, 2, 33},
		},
	}

	qry, err := query.NewBuilder().AndContains("Height", "2").Query()
	require.NoError(t, err)
	match := qry.Matches(be)
	require.True(t, match)
}
