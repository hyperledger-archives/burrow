package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	"github.com/tendermint/go-wire"
)

type RebondTx struct {
	Address   crypto.Address
	Height    int
	Signature crypto.Signature
	txHashMemoizer
}

var _ Tx = &RebondTx{}

func NewRebondTx(addr crypto.Address, height int) *RebondTx {
	return &RebondTx{
		Address: addr,
		Height:  height,
	}
}

func (tx *RebondTx) Sign(chainID string, signingAccounts ...acm.AddressableSigner) error {
	if len(signingAccounts) != 1 {
		return fmt.Errorf("RebondTx expects a single AddressableSigner for its signature but %v were provieded",
			len(signingAccounts))
	}
	var err error
	tx.Signature, err = crypto.ChainSign(signingAccounts[0], chainID, tx)
	if err != nil {
		return fmt.Errorf("could not sign %v: %v", tx, err)
	}
	return nil
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
