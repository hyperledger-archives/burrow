package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/tendermint/go-wire"
)

type CallTx struct {
	Input TxInput `json:"input"`
	// Pointer since CallTx defines unset 'to' address as inducing account creation
	Address  *acm.Address `json:"address"`
	GasLimit uint64       `json:"gas_limit"`
	Fee      uint64       `json:"fee"`
	Data     []byte       `json:"data"`
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

	return &CallTx{
		Input: TxInput{
			Address:   from.Address(),
			Amount:    amt,
			Sequence:  sequence,
			PublicKey: from,
		},
		Address:  to,
		GasLimit: gasLimit,
		Fee:      fee,
		Data:     data,
	}
}

func (tx *CallTx) Sign(chainID string, signingAccounts ...acm.AddressableSigner) error {
	if len(signingAccounts) != 1 {
		return fmt.Errorf("CallTx expects a single AddressableSigner for its single Input but %v were provieded",
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
	signJson := fmt.Sprintf(`{"chain_id":%s,"tx":[%v,{"address":"%s","data":"%X","fee":%v,"gas_limit":%v,"input":%s}]}`,
		jsonEscape(chainID), TxTypeCall, tx.Address, tx.Data, tx.Fee, tx.GasLimit, tx.Input.SignString())

	wire.WriteTo([]byte(signJson), w, n, err)
}

func (tx *CallTx) GetInputs() []TxInput {
	return []TxInput{tx.Input}
}

func (tx *CallTx) String() string {
	return fmt.Sprintf("CallTx{%v -> %s: %X}", tx.Input, tx.Address, tx.Data)
}

func (tx *CallTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}
