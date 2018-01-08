package event

import (
	"encoding/json"
	"testing"

	acm "github.com/hyperledger/burrow/account"
	exe_events "github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	tm_types "github.com/tendermint/tendermint/types"
)

func TestSerialiseTMEventData(t *testing.T) {
	roundTripAnyEventData(t, AnyEventData{
		TMEventData: &tm_types.TMEventData{
			TMEventDataInner: tm_types.EventDataNewBlock{
				Block: &tm_types.Block{
					LastCommit: &tm_types.Commit{},
					Header: &tm_types.Header{
						ChainID: "ChainID-ChainEgo",
					},
					Data: &tm_types.Data{},
				},
			},
		},
	})

}

func TestSerialiseEVMEventData(t *testing.T) {
	roundTripAnyEventData(t, AnyEventData{
		BurrowEventData: &EventData{
			EventDataInner: exe_events.EventDataTx{
				Tx: &txs.CallTx{
					Address: &acm.Address{1, 2, 2, 3},
				},
				Return:    []byte{1, 2, 3},
				Exception: "Exception",
			},
		},
	})
}

func TestSerialiseError(t *testing.T) {
	s := "random error"
	roundTripAnyEventData(t, AnyEventData{
		Err: &s,
	})
}

func roundTripAnyEventData(t *testing.T, aed AnyEventData) {
	bs, err := json.Marshal(aed)
	assert.NoError(t, err)

	aedOut := new(AnyEventData)
	err = json.Unmarshal(bs, aedOut)
	assert.NoError(t, err)

	bsOut, err := json.Marshal(aedOut)
	assert.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))

}
