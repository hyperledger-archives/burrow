package types

import (
	tendermint_consensus "github.com/tendermint/tendermint/consensus"
	tendermint_types "github.com/tendermint/tendermint/types"
)

// ConsensusState
type ConsensusState struct {
	Height     int                        `json:"height"`
	Round      int                        `json:"round"`
	Step       uint8                      `json:"step"`
	StartTime  string                     `json:"start_time"`
	CommitTime string                     `json:"commit_time"`
	Validators []Validator                `json:"validators"`
	Proposal   *tendermint_types.Proposal `json:"proposal"`
}

func FromRoundState(rs *tendermint_consensus.RoundState) *ConsensusState {
	cs := &ConsensusState{
		CommitTime: rs.CommitTime.String(),
		Height:     rs.Height,
		Proposal:   rs.Proposal,
		Round:      rs.Round,
		StartTime:  rs.StartTime.String(),
		Step:       uint8(rs.Step),
		Validators: FromTendermintValidators(rs.Validators.Validators),
	}
	return cs
}
