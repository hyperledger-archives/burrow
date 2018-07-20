package payload

import (
	"fmt"
)

func (txIn *TxInput) String() string {
	return fmt.Sprintf("TxInput{%s, Amount: %v, Sequence:%v}", txIn.Address, txIn.Amount, txIn.Sequence)
}
