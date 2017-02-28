// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
