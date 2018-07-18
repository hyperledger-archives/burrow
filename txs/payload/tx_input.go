package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

func (txIn *TxInput) ValidateBasic() error {
	if txIn.Address == crypto.ZeroAddress {
		return ErrTxInvalidAddress
	}
	if txIn.Amount == 0 {
		return ErrTxInvalidAmount
	}
	return nil
}

func (txIn *TxInput) String() string {
	return fmt.Sprintf("TxInput{%s, Amount: %v, Sequence:%v}", txIn.Address, txIn.Amount, txIn.Sequence)
}
