package validator

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
)

func (v Validator) String() string {
	return fmt.Sprintf("Validator{Address: %v, PublicKey: %v, Power: %v}", v.Address, v.PublicKey, v.Power)
}

func (v Validator) FillAddress() {
	if v.Address == nil {
		address := v.PublicKey.GetAddress()
		v.Address = &address
	}
}

func FromAccount(acc *acm.Account, power uint64) Validator {
	address := acc.GetAddress()
	return Validator{
		Address:   &address,
		PublicKey: acc.GetPublicKey(),
		Power:     power,
	}
}
