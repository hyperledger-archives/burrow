package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/go-wire"
)

type BondTx struct {
	From   TxInput       `json:"from_account"`
	To     TxOutput      `json:"to_validator"`
	BondTo acm.PublicKey `json:"bond_to"`
	txHashMemoizer
}

var _ Tx = &BondTx{}

func NewBondTx(from acm.Address, to acm.PublicKey, amount uint64, sequence uint64, fee uint64) (*BondTx, error) {
	return &BondTx{
		From: TxInput{
			Address:  from,
			Amount:   amount + fee,
			Sequence: sequence,
		},
		To: TxOutput{
			Address: to.Address(),
			Amount:  amount,
		},
		BondTo: to,
	}, nil
}

func (tx *BondTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	signJson := fmt.Sprintf(`{"chain_id":%s,"tx":[%v,{"bond_to":"%v","from":"%v","to":%s}]}`,
		jsonEscape(chainID), TxTypeBond, tx.BondTo, tx.From.SignString(), tx.To.SignString())

	wire.WriteTo([]byte(signJson), w, n, err)
}

func (tx *BondTx) GetInputs() []TxInput {
	return []TxInput{tx.From}
}

func (tx *BondTx) String() string {
	return fmt.Sprintf("BondTx{%v: %v -> %v}", tx.From.Amount, tx.From.Address, tx.To.Address)
}

func (tx *BondTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}

func (tx *BondTx) Sign(chainID string, signingAccounts ...acm.AddressableSigner) error {
	if len(signingAccounts) != 1 {
		return fmt.Errorf("BondTx expects a single AddressableSigner for its single Input but %v were provieded",
			len(signingAccounts))
	}
	var err error
	tx.From.PublicKey = signingAccounts[0].PublicKey()
	tx.From.Signature, err = acm.ChainSign(signingAccounts[0], chainID, tx)
	if err != nil {
		return fmt.Errorf("could not sign %v: %v", tx, err)
	}
	return nil
}
