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

package tendermint

import (
	tendermint_types "github.com/tendermint/tendermint/types"

	consensus_types "github.com/eris-ltd/eris-db/consensus/types"
)

// Anticipating moving to our own definition of Validator, or at least
// augmenting Tendermint's.
type TendermintValidator struct {
	tmintValidator *tendermint_types.Validator `json:"validator"`
}

var _ consensus_types.Validator = (*TendermintValidator)(nil)

func (tendermintValidator *TendermintValidator) AssertIsValidator() {

}

func (tendermintValidator *TendermintValidator) Address() []byte {
	return tendermintValidator.tmintValidator.Address
}

//-------------------------------------------------------------------------------------
// Helper function for TendermintValidator

func FromTendermintValidators(tmValidators []*tendermint_types.Validator) []Validator {
	validators := make([]Validator, len(tmValidators))
	for i, tmValidator := range tmValidators {
		// This embedding could be replaced by a mapping once if we want to describe
		// a more general notion of validator
		validators[i] = &TendermintValidator{tmintValidator: tmValidator}
	}
	return validators
}
