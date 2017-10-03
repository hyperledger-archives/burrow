package account

import (
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

type Validator interface {
	Addressable
	// The validator's voting power
	Power() uint64
	// Alter the validator's voting power by amount that can be negative or positive.
	// A power of 0 effectively unbonds the validator
	AlterPower(amount int64) Validator
}

// Neither abci_types or tm_types has quite the representation we want
type ConcreteValidator struct {
	Address Address       `json:"address"`
	PubKey  crypto.PubKey `json:"pub_key"`
	Power   uint64        `json:"power"`
}

type concreteValidatorWrapper struct {
	*ConcreteValidator `json:"unwrap"`
}

var _ Validator = concreteValidatorWrapper{}

var _ = wire.RegisterInterface(struct{ Validator }{}, wire.ConcreteType{concreteValidatorWrapper{}, 0x01})

func (cvw concreteValidatorWrapper) Address() Address {
	return cvw.ConcreteValidator.Address
}

func (cvw concreteValidatorWrapper) PubKey() crypto.PubKey {
	return cvw.ConcreteValidator.PubKey
}

func (cvw concreteValidatorWrapper) Power() uint64 {
	return cvw.ConcreteValidator.Power
}

func (cvw concreteValidatorWrapper) AlterPower(amount int64) Validator {
	//cv := *cvw.ConcreteValidator
	//cv.Power = cv.Power + amount
	//cvw.ConcreteValidator.Power
	return cvw
}

func (cv ConcreteValidator) Validator() Validator {
	return concreteValidatorWrapper{&cv}
}

func (cv *ConcreteValidator) Copy() *ConcreteValidator {
	cvCopy := *cv
	return &cvCopy
}
