package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/go-wire"
)

type TxInput struct {
	Address   acm.Address
	Amount    uint64
	Sequence  uint64
	Signature acm.Signature
	PublicKey acm.PublicKey
}

func (txIn *TxInput) ValidateBasic() error {
	if len(txIn.Address) != 20 {
		return ErrTxInvalidAddress
	}
	if txIn.Amount == 0 {
		return ErrTxInvalidAmount
	}
	return nil
}

func (txIn *TxInput) WriteSignBytes(w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"address":"%s","amount":%v,"sequence":%v}`, txIn.Address, txIn.Amount, txIn.Sequence)), w, n, err)
}

func (txIn *TxInput) String() string {
	return fmt.Sprintf("TxInput{%s,%v,%v,%v,%v}", txIn.Address, txIn.Amount, txIn.Sequence, txIn.Signature, txIn.PublicKey)
}
