package txs

import (
	"fmt"
	"github.com/hyperledger/burrow/crypto"
	"bytes"
)

type TxInput struct {
	Address   crypto.Address
	PublicKey crypto.PublicKey
	Signature crypto.Signature
	Amount    uint64
	Sequence  uint64
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

func (txIn *TxInput) SignBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf(`{"address":"%s","amount":%v,"sequence":%v}`, txIn.Address, txIn.Amount, txIn.Sequence))
	return buf.Bytes(), nil
}

func (txIn *TxInput) String() string {
	return fmt.Sprintf("TxInput{%s,%v,%v,%v,%v}", txIn.Address, txIn.Amount, txIn.Sequence, txIn.Signature, txIn.PublicKey)
}
