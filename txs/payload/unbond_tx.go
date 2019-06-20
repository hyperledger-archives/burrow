package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

func NewUnbondTx(address crypto.Address, height uint64) *UnbondTx {
	return &UnbondTx{
		Input:  &TxInput{},
		Output: &TxOutput{},
	}
}

func (tx *UnbondTx) Type() Type {
	return TypeUnbond
}

func (tx *UnbondTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *UnbondTx) String() string {
	return fmt.Sprintf("UnbondTx{%v -> %v}", tx.Input.Address, tx.Output.Address)
}

func (tx *UnbondTx) Any() *Any {
	return &Any{
		UnbondTx: tx,
	}
}
