package txs

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
)

type TxOutput struct {
	Address acm.Address `json:"address"`
	Amount  uint64      `json:"amount"`
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

func (txOut *TxOutput) SignString() string {
	return fmt.Sprintf(`{"address":"%s","amount":%v}`, txOut.Address, txOut.Amount)
}

func (txOut *TxOutput) String() string {
	return fmt.Sprintf("TxOutput{%s,%v}", txOut.Address, txOut.Amount)
}
