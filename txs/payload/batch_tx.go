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

func (item *BatchTxItem) String() string {
	tx := item.GetValue()
	switch tx := tx.(type) {
	case CallTx:
		return tx.String()
	case SendTx:
		return tx.String()
	case NameTx:
		return tx.String()
	case PermsTx:
		return tx.String()
	case GovTx:
		return tx.String()
	case BondTx:
		return tx.String()
	case UnbondTx:
		return tx.String()
	default:
		panic(fmt.Sprintf("unknown item %v in BatchTx", tx))
	}
}

func (item *BatchTxItem) GetPayload() Payload {
	if item.CallTx != nil {
		return item.CallTx
	} else if item.SendTx != nil {
		return item.SendTx
	} else if item.NameTx != nil {
		return item.NameTx
	} else if item.BondTx != nil {
		return item.BondTx
	} else if item.UnbondTx != nil {
		return item.UnbondTx
	} else {
		panic(fmt.Sprintf("unknown item %v in BatchTx", item))
	}
}
