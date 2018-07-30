package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/execution/errors"
)

func (input *TxInput) String() string {
	return fmt.Sprintf("TxInput{%s, Amount: %v, Sequence:%v}", input.Address, input.Amount, input.Sequence)
}

func (input *TxInput) Validate(acc acm.Account) error {
	if input.Address != acc.Address() {
		return fmt.Errorf("trying to validate input from address %v but passed account %v", input.Address,
			acc.Address())
	}
	// Check sequences
	if acc.Sequence()+1 != uint64(input.Sequence) {
		return ErrTxInvalidSequence{
			Input:   input,
			Account: acc,
		}
	}
	// Check amount
	if acc.Balance() < uint64(input.Amount) {
		return errors.ErrorCodeInsufficientFunds
	}
	return nil
}

func ValidateInputs(getter state.AccountGetter, ins []*TxInput) error {
	for _, in := range ins {
		acc, err := getter.GetAccount(in.Address)
		if err != nil {
			return err
		}
		if acc == nil {
			return fmt.Errorf("validateInputs() expects to be able to retrive accoutn %v but it was not found",
				in.Address)
		}
		err = in.Validate(acc)
		if err != nil {
			return err
		}
	}
	return nil
}
