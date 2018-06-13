package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

type TxInput struct {
	Address  crypto.Address
	Amount   uint64
	Sequence uint64
}

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
	return fmt.Sprintf("TxInput{%s, Amt: %v, Seq:%v}", txIn.Address, txIn.Amount, txIn.Sequence)
}
