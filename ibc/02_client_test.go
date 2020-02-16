package ibc

import (
	"testing"

	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func TestCreate(t *testing.T) {
	state := tendermint.ConsensusState{
		Root:             commitment.NewRoot([]byte("00000")),
		ValidatorSetHash: []byte("00000"),
	}

	NewClientCreate(state, "12345", []byte("00000"))
	// require.NoError(t, err)
	// data, err := json.Marshal(msg)
	// require.NoError(t, err)

	// fmt.Println(string(data))
}
