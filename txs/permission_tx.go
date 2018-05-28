package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/permission/snatives"
	"github.com/tendermint/go-wire"
)

type PermissionsTx struct {
	Input    TxInput           `json:"input"`
	PermArgs snatives.PermArgs `json:"args`
	txHashMemoizer
}

var _ Tx = &PermissionsTx{}

func NewPermissionsTx(st state.AccountGetter, from acm.PublicKey, args snatives.PermArgs) (*PermissionsTx, error) {
	addr := from.Address()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("Invalid address %s from pubkey %s", addr, from)
	}

	sequence := acc.Sequence() + 1
	return NewPermissionsTxWithSequence(from, args, sequence), nil
}

func NewPermissionsTxWithSequence(from acm.PublicKey, args snatives.PermArgs, sequence uint64) *PermissionsTx {

	return &PermissionsTx{
		Input: TxInput{
			Address:   from.Address(),
			Amount:    1, // NOTE: amounts can't be 0 ...
			Sequence:  sequence,
			PublicKey: from,
		},
		PermArgs: args,
	}
}

func (tx *PermissionsTx) Sign(chainID string, signingAccounts ...acm.AddressableSigner) error {
	if len(signingAccounts) != 1 {
		return fmt.Errorf("PermissionsTx expects a single AddressableSigner for its single Input but %v were provieded",
			len(signingAccounts))
	}
	var err error
	tx.Input.PublicKey = signingAccounts[0].PublicKey()
	tx.Input.Signature, err = acm.ChainSign(signingAccounts[0], chainID, tx)
	if err != nil {
		return fmt.Errorf("could not sign %v: %v", tx, err)
	}
	return nil
}

func (tx *PermissionsTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	signJson := fmt.Sprintf(`{"chain_id":%s,"tx":[%v,{"args":%s,"input":%s}]}`,
		jsonEscape(chainID), TxTypePermissions, tx.PermArgs, tx.Input.SignString())

	wire.WriteTo([]byte(signJson), w, n, err)
}

func (tx *PermissionsTx) GetInputs() []TxInput {
	return []TxInput{tx.Input}
}

func (tx *PermissionsTx) String() string {
	return fmt.Sprintf("PermissionsTx{%v -> %v}", tx.Input, tx.PermArgs)
}

func (tx *PermissionsTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}
