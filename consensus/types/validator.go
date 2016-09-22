package types

import (
	"github.com/tendermint/go-wire"
	tendermint_types "github.com/tendermint/tendermint/types"
)

var _ = wire.RegisterInterface(
	struct{ Validator }{},
	wire.ConcreteType{&TendermintValidator{}, byte(0x01)},
)

type Validator interface {
	AssertIsValidator()
	Address() []byte
}

// Anticipating moving to our own definition of Validator, or at least
// augmenting Tendermint's.
type TendermintValidator struct {
	*tendermint_types.Validator `json:"validator"`
}

var _ Validator = (*TendermintValidator)(nil)

func (tendermintValidator *TendermintValidator) AssertIsValidator() {

}

func (tendermintValidator *TendermintValidator) Address() []byte {
	return tendermintValidator.Address()
}

//-------------------------------------------------------------------------------------
// Helper function for TendermintValidator

func FromTendermintValidators(tmValidators []*tendermint_types.Validator) []Validator {
	validators := make([]Validator, len(tmValidators))
	for i, tmValidator := range tmValidators {
		// This embedding could be replaced by a mapping once if we want to describe
		// a more general notion of validator
		validators[i] = &TendermintValidator{tmValidator}
	}
	return validators
}
