package types

import (
	"time"

	tendermint_consensus "github.com/tendermint/tendermint/consensus"
	tendermint_types "github.com/tendermint/tendermint/types"
)

// ConsensusState
type ConsensusState struct {
	Height     int                        `json:"height"`
	Round      int                        `json:"round"`
	Step       uint8                      `json:"step"`
	StartTime  time.Time                  `json:"start_time"`
	CommitTime time.Time                  `json:"commit_time"`
	Validators []Validator                `json:"validators"`
	Proposal   *tendermint_types.Proposal `json:"proposal"`
}

func FromRoundState(rs *tendermint_consensus.RoundState) *ConsensusState {
	cs := &ConsensusState{
		StartTime:  rs.StartTime,
		CommitTime: rs.CommitTime,
		Height:     rs.Height,
		Proposal:   rs.Proposal,
		Round:      rs.Round,
		Step:       uint8(rs.Step),
		Validators: FromTendermintValidators(rs.Validators.Validators),
	}
	return cs
}
