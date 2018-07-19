package payload

import (
	"fmt"
)

func (txOut *TxOutput) ValidateBasic() error {
	if len(txOut.Address) != 20 {
		return ErrTxInvalidAddress
	}
	if txOut.Amount == 0 {
		return ErrTxInvalidAmount
	}
	return nil
}

func (txOut *TxOutput) String() string {
	return fmt.Sprintf("TxOutput{%s, Amount: %v}", txOut.Address, txOut.Amount)
}
