package payload

import (
	"fmt"
)

func (tx *BatchTx) Type() Type {
	return TypeBatch
}

func (tx *BatchTx) GetInputs() []*TxInput {
	return make([]*TxInput, 0)
}

func (tx *BatchTx) String() string {
	return fmt.Sprintf("BatchTx{%v}", tx.Txs)
}

func (tx *BatchTx) Any() *Any {
	return &Any{
		BatchTx: tx,
	}
}
