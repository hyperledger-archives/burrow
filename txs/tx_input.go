package txs

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
)

type TxInput struct {
	Address   acm.Address   `json:"address"`
	Amount    uint64        `json:"amount"`
	Sequence  uint64        `json:"sequence"`
	Signature acm.Signature `json:"signature"`
	PublicKey acm.PublicKey `json:"public_key"`
}

func (txIn *TxInput) ValidateBasic() error {
	if len(txIn.Address) != 20 {
		return ErrTxInvalidAddress
	}

	if txIn.Address != txIn.PublicKey.Address() {
		return ErrTxInvalidAddress
	}

	if txIn.Amount == 0 {
		return ErrTxInvalidAmount
	}
	return nil
}

func (txIn *TxInput) SignString() string {
	return fmt.Sprintf(`{"address":"%s","amount":%v,"sequence":%v}`,
		txIn.Address, txIn.Amount, txIn.Sequence)
}

func (txIn *TxInput) String() string {
	return fmt.Sprintf("TxInput{%s,%v,%v,%v,%v}", txIn.PublicKey.Address(), txIn.Amount, txIn.Sequence, txIn.Signature, txIn.PublicKey)
}
