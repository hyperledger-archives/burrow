package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/tendermint/go-wire"
)

type CallTx struct {
	Input *TxInput
	// Pointer since CallTx defines unset 'to' address as inducing account creation
	Address  *acm.Address
	GasLimit uint64
	Fee      uint64
	Data     []byte
	txHashMemoizer
}

var _ Tx = &CallTx{}

func NewCallTx(st state.AccountGetter, from acm.PublicKey, to *acm.Address, data []byte,
	amt, gasLimit, fee uint64) (*CallTx, error) {

	addr := from.Address()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("invalid address %s from pubkey %s", addr, from)
	}

	sequence := acc.Sequence() + 1
	return NewCallTxWithSequence(from, to, data, amt, gasLimit, fee, sequence), nil
}

func NewCallTxWithSequence(from acm.PublicKey, to *acm.Address, data []byte,
	amt, gasLimit, fee, sequence uint64) *CallTx {
	input := &TxInput{
		Address:   from.Address(),
		Amount:    amt,
		Sequence:  sequence,
		PublicKey: from,
	}

	return &CallTx{
		Input:    input,
		Address:  to,
		GasLimit: gasLimit,
		Fee:      fee,
		Data:     data,
	}
}

func (tx *CallTx) Sign(chainID string, signingAccounts ...acm.SigningAccount) error {
	if len(signingAccounts) != 1 {
		return fmt.Errorf("CallTx expects a single SigningAccount for its single Input but %v were provieded",
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

func (tx *CallTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s`, jsonEscape(chainID))), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"tx":[%v,{"address":"%s","data":"%X"`, TxTypeCall, tx.Address, tx.Data)), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"fee":%v,"gas_limit":%v,"input":`, tx.Fee, tx.GasLimit)), w, n, err)
	tx.Input.WriteSignBytes(w, n, err)
	wire.WriteTo([]byte(`}]}`), w, n, err)
}

func (tx *CallTx) GetInputs() []TxInput {
	return []TxInput{*tx.Input}
}

func (tx *CallTx) String() string {
	return fmt.Sprintf("CallTx{%v -> %s: %X}", tx.Input, tx.Address, tx.Data)
}

func (tx *CallTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}
