// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

// Consensus is part of the pipe for ErisMint and provides the implementation
// for the pipe to call into the ErisMint application

package erismint
import (
	"github.com/tendermint/tendermint/types"

  core_types "github.com/eris-ltd/eris-db/core/types"
)

// The consensus struct.
type consensus struct {
	erisMint *ErisMint
}

func newConsensus(erisMint *ErisMint) *consensus {
	return &consensus{erisMint}
}

// Get the current consensus state.
func (this *consensus) State() (*core_types.ConsensusState, error) {
	// TODO-RPC!
	return &core_types.ConsensusState{}, nil
}

// Get all validators.
func (this *consensus) Validators() (*core_types.ValidatorList, error) {
	var blockHeight int
	bondedValidators := make([]*types.Validator, 0)
	unbondingValidators := make([]*types.Validator, 0)

	s := this.erisMint.GetState()
	blockHeight = s.LastBlockHeight

	// TODO: rpc

	/*
		s.BondedValidators.Iterate(func(index int, val *types.Validator) bool {
			bondedValidators = append(bondedValidators, val)
			return false
		})
		s.UnbondingValidators.Iterate(func(index int, val *types.Validator) bool {
			unbondingValidators = append(unbondingValidators, val)
			return false
		})*/

	return &core_types.ValidatorList{blockHeight, bondedValidators,
    unbondingValidators}, nil
}
