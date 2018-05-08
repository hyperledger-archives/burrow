package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/go-wire"
)

type TxOutput struct {
	Address acm.Address
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

func (txOut *TxOutput) WriteSignBytes(w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"address":"%s","amount":%v}`, txOut.Address, txOut.Amount)), w, n, err)
}

func (txOut *TxOutput) String() string {
	return fmt.Sprintf("TxOutput{%s,%v}", txOut.Address, txOut.Amount)
}
