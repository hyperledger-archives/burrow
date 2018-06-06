package txs

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

type UnbondTx struct {
	Address   crypto.Address
	Height    int
	Signature crypto.Signature
	txHashMemoizer
}

var _ Tx = &UnbondTx{}

func NewUnbondTx(addr crypto.Address, height int) *UnbondTx {
	return &UnbondTx{
		Address: addr,
		Height:  height,
	}
}

func (tx *UnbondTx) Type() TxType {
	return TxTypeUnbond
}

func (tx *UnbondTx) GetInputs() []TxInput {
	return nil
}

func (tx *UnbondTx) String() string {
	return fmt.Sprintf("UnbondTx{%s,%v,%v}", tx.Address, tx.Height, tx.Signature)
}
