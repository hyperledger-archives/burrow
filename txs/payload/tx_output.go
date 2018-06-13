package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

type TxOutput struct {
	Address crypto.Address
	Amount  uint64
}

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
	return fmt.Sprintf("TxOutput{%s,%v}", txOut.Address, txOut.Amount)
}
