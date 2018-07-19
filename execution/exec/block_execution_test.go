package exec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockExecution_Marshal(t *testing.T) {
	be := &BlockExecution{
		// TODO: reenable when tendermint Header works GRPC
		//BlockHeader: &abciTypes.Header{
		//	Height:  3,
		//	AppHash: []byte{2},
		//	Proposer: abciTypes.Validator{
		//		Power: 34,
		//	},
		//},
	}
	bs, err := be.Marshal()
	require.NoError(t, err)
	beOut := new(BlockExecution)
	require.NoError(t, beOut.Unmarshal(bs))
}
