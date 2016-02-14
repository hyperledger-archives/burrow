package pipe

import (
	"github.com/tendermint/tendermint/types"

	"github.com/eris-ltd/eris-db/tmsp"
)

// The consensus struct.
type consensus struct {
	erisdbApp *tmsp.ErisDBApp
}

func newConsensus(erisdbApp *tmsp.ErisDBApp) *consensus {
	return &consensus{erisdbApp}
}

// Get the current consensus state.
func (this *consensus) State() (*ConsensusState, error) {
	// TODO-RPC!
	return &ConsensusState{}, nil
}

// Get all validators.
func (this *consensus) Validators() (*ValidatorList, error) {
	var blockHeight int
	bondedValidators := make([]*types.Validator, 0)
	unbondingValidators := make([]*types.Validator, 0)

	s := this.erisdbApp.GetState()
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

	return &ValidatorList{blockHeight, bondedValidators, unbondingValidators}, nil
}
