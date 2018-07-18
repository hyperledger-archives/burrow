package payload

import (
	"fmt"
)

func (tx *GovernanceTx) Type() Type {
	return TypeGovernance
}

func (tx *GovernanceTx) GetInputs() []*TxInput {
	return tx.Inputs
}

func (tx *GovernanceTx) String() string {
	return fmt.Sprintf("GovernanceTx{%v -> %v}", tx.Inputs, tx.AccountUpdates)
}
