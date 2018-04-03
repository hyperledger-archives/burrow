package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/go-wire"
)

type UnbondTx struct {
	Address   acm.Address
	Height    int
	Signature acm.Signature
	txHashMemoizer
}

var _ Tx = &UnbondTx{}

func NewUnbondTx(addr acm.Address, height int) *UnbondTx {
	return &UnbondTx{
		Address: addr,
		Height:  height,
	}
}

func (tx *UnbondTx) Sign(chainID string, privAccount acm.SigningAccount) {
	tx.Signature = acm.ChainSign(privAccount, chainID, tx)
}

func (tx *UnbondTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s`, jsonEscape(chainID))), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"tx":[%v,{"address":"%s","height":%v}]}`, TxTypeUnbond, tx.Address, tx.Height)), w, n, err)
}

func (tx *UnbondTx) GetInputs() []TxInput {
	return nil
}

func (tx *UnbondTx) String() string {
	return fmt.Sprintf("UnbondTx{%s,%v,%v}", tx.Address, tx.Height, tx.Signature)
}

func (tx *UnbondTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}
