package payload

import (
	"fmt"
)

func NewInterchainTx() *InterchainTx {
	return &InterchainTx{}
}

func (tx *InterchainTx) Type() Type {
	return TypeInterchain
}

func (tx *InterchainTx) GetInputs() []*TxInput {
	return nil
}

func (tx *InterchainTx) String() string {
	return fmt.Sprintf("InterchainTx{}")
}

func (tx *InterchainTx) Any() *Any {
	return &Any{
		InterchainTx: tx,
	}
}
