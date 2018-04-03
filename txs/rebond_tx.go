package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/go-wire"
)

type RebondTx struct {
	Address   acm.Address
	Height    int
	Signature acm.Signature
	txHashMemoizer
}

var _ Tx = &RebondTx{}

func NewRebondTx(addr acm.Address, height int) *RebondTx {
	return &RebondTx{
		Address: addr,
		Height:  height,
	}
}

func (tx *RebondTx) Sign(chainID string, privAccount acm.SigningAccount) {
	tx.Signature = acm.ChainSign(privAccount, chainID, tx)
}

func (tx *RebondTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s`, jsonEscape(chainID))), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"tx":[%v,{"address":"%s","height":%v}]}`, TxTypeRebond, tx.Address, tx.Height)), w, n, err)
}

func (tx *RebondTx) GetInputs() []TxInput {
	return nil
}

func (tx *RebondTx) String() string {
	return fmt.Sprintf("RebondTx{%s,%v,%v}", tx.Address, tx.Height, tx.Signature)
}

func (tx *RebondTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}
