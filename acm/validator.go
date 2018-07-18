package acm

import (
	"encoding/json"

	"github.com/hyperledger/burrow/crypto"
)

type Validator interface {
	crypto.Addressable
	// The validator's voting power
	Power() uint64
}

// Neither abci_types or tm_types has quite the representation we want
type ConcreteValidator struct {
	Address   crypto.Address
	PublicKey crypto.PublicKey
	Power     uint64
}

type concreteValidatorWrapper struct {
	*ConcreteValidator `json:"unwrap"`
}

var _ Validator = concreteValidatorWrapper{}

func AsValidator(account Account) Validator {
	return ConcreteValidator{
		Address:   account.Address(),
		PublicKey: account.PublicKey(),
		Power:     account.Balance(),
	}.Validator()
}

func AsConcreteValidator(validator Validator) *ConcreteValidator {
	if validator == nil {
		return nil
	}
	if ca, ok := validator.(concreteValidatorWrapper); ok {
		return ca.ConcreteValidator
	}
	return &ConcreteValidator{
		Address:   validator.Address(),
		PublicKey: validator.PublicKey(),
		Power:     validator.Power(),
	}
}

func (cvw concreteValidatorWrapper) Address() crypto.Address {
	return cvw.ConcreteValidator.Address
}

func (cvw concreteValidatorWrapper) PublicKey() crypto.PublicKey {
	return cvw.ConcreteValidator.PublicKey
}

func (cvw concreteValidatorWrapper) Power() uint64 {
	return cvw.ConcreteValidator.Power
}

func (cvw concreteValidatorWrapper) WithNewPower(power uint64) Validator {
	cv := cvw.Copy()
	cv.Power = power
	return concreteValidatorWrapper{
		ConcreteValidator: cv,
	}
}

func (cv ConcreteValidator) Validator() Validator {
	return concreteValidatorWrapper{
		ConcreteValidator: &cv,
	}
}

func (cv *ConcreteValidator) Copy() *ConcreteValidator {
	cvCopy := *cv
	return &cvCopy
}

func (cv *ConcreteValidator) String() string {
	if cv == nil {
		return "Nil Validator"
	}

	bs, err := json.Marshal(cv)
	if err != nil {
		return "error serialising Validator"
	}

	return string(bs)
}
