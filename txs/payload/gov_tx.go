package payload

import (
	"fmt"
)

func (tx *GovTx) Type() Type {
	return TypeGovernance
}

func (tx *GovTx) GetInputs() []*TxInput {
	return tx.Inputs
}

func (tx *GovTx) String() string {
	return fmt.Sprintf("GovTx{%v -> %v}", tx.Inputs, tx.AccountUpdates)
}

func (tx *GovTx) Any() *Any {
	return &Any{
		GovTx: tx,
	}
}
