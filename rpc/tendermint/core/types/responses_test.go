package types

import (
	"testing"

	"time"

	consensus_types "github.com/eris-ltd/eris-db/consensus/types"
	"github.com/tendermint/go-wire"
	tendermint_types "github.com/tendermint/tendermint/types"
)

func TestResultDumpConsensusState(t *testing.T) {
	result := ResultDumpConsensusState{
		ConsensusState: &consensus_types.ConsensusState{
			Height:     3,
			Round:      1,
			Step:       uint8(1),
			StartTime:  time.Now().Add(-time.Second * 100),
			CommitTime: time.Now().Add(-time.Second * 10),
			Validators: []consensus_types.Validator{
				&consensus_types.TendermintValidator{},
			},
			Proposal: &tendermint_types.Proposal{},
		},
		PeerConsensusStates: []*ResultPeerConsensusState{
			{
				PeerKey:            "Foo",
				PeerConsensusState: "Bar",
			},
		},
	}
	wire.JSONBytes(result)
}
