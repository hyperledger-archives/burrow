package event

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	tm_types "github.com/tendermint/tendermint/types"
)

func TestSerialise(t *testing.T) {
	tmEventData := tm_types.TMEventData{
		TMEventDataInner: tm_types.EventDataNewBlock{
			Block: &tm_types.Block{
				LastCommit: &tm_types.Commit{},
				Header: &tm_types.Header{
					ChainID: "ChainID-ChainEgo",
				},
				Data: &tm_types.Data{},
			},
		},
	}
	aed := AnyEventData{
		TMEventData: &tmEventData,
	}

	bs, err := json.Marshal(aed)
	assert.NoError(t, err)

	aedOut := new(AnyEventData)
	err = json.Unmarshal(bs, aedOut)
	assert.NoError(t, err)

	assert.Equal(t, aed.TMEventData.Unwrap().(tm_types.EventDataNewBlock).Block.ChainID,
		aedOut.TMEventData.Unwrap().(tm_types.EventDataNewBlock).Block.ChainID)

	bsOut, err := json.Marshal(aed)
	assert.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}
