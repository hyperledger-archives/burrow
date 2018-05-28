package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/go-wire"
)

type UnbondTx struct {
	From TxInput  `json:"from_validator"`
	To   TxOutput `json:"to_account"`
	txHashMemoizer
}

var _ Tx = &UnbondTx{}

func NewUnbondTx(from, to acm.Address, amount uint64, sequence uint64, fee uint64) (*UnbondTx, error) {
	return &UnbondTx{
		From: TxInput{
			Address:  from,
			Amount:   amount + fee,
			Sequence: sequence,
		},
		To: TxOutput{
			Address: to,
			Amount:  amount,
		},
	}, nil
}

func (tx *UnbondTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	signJson := fmt.Sprintf(`{"chain_id":%s,"tx":[%v,{"from":"%v","to":%s}]}`,
		jsonEscape(chainID), TxTypeUnbond, tx.From.SignString(), tx.To.SignString())

	wire.WriteTo([]byte(signJson), w, n, err)
}

func (tx *UnbondTx) GetInputs() []TxInput {
	return []TxInput{tx.From}
}

func (tx *UnbondTx) String() string {
	return fmt.Sprintf("UnbondTx{%v: %v -> %v}", tx.From.Amount, tx.From.Address, tx.To.Address)
}

func (tx *UnbondTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}

func (tx *UnbondTx) Sign(chainID string, signingAccounts ...acm.AddressableSigner) error {
	if len(signingAccounts) != 1 {
		return fmt.Errorf("UnbondTx expects a single AddressableSigner for its signature but %v were provided",
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
