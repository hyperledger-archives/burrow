package payload

import (
	"fmt"
)

func (tx *IdentifyTx) Type() Type {
	return TypeIdentify
}

func (tx *IdentifyTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *IdentifyTx) String() string {
	return fmt.Sprintf("IdentifyTx{%v -> %v}", tx.Input, tx.Node.NetAddress)
}

func (tx *IdentifyTx) Any() *Any {
	return &Any{
		IdentifyTx: tx,
	}
}
