package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

type UnbondTx struct {
	Input   *TxInput
	Address crypto.Address
	Height  int
}

func NewUnbondTx(addr crypto.Address, height int) *UnbondTx {
	return &UnbondTx{
		Address: addr,
		Height:  height,
	}
}

func (tx *UnbondTx) Type() Type {
	return TypeUnbond
}

func (tx *UnbondTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *UnbondTx) String() string {
	return fmt.Sprintf("UnbondTx{%v -> %s,%v}", tx.Input, tx.Address, tx.Height)
}
