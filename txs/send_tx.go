package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/tendermint/go-wire"
)

type SendTx struct {
	Inputs  []*TxInput
	Outputs []*TxOutput
	txHashMemoizer
}

var _ Tx = &SendTx{}

func NewSendTx() *SendTx {
	return &SendTx{
		Inputs:  []*TxInput{},
		Outputs: []*TxOutput{},
	}
}

func (tx *SendTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s`, jsonEscape(chainID))), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"tx":[%v,{"inputs":[`, TxTypeSend)), w, n, err)
	for i, in := range tx.Inputs {
		in.WriteSignBytes(w, n, err)
		if i != len(tx.Inputs)-1 {
			wire.WriteTo([]byte(","), w, n, err)
		}
	}
	wire.WriteTo([]byte(`],"outputs":[`), w, n, err)
	for i, out := range tx.Outputs {
		out.WriteSignBytes(w, n, err)
		if i != len(tx.Outputs)-1 {
			wire.WriteTo([]byte(","), w, n, err)
		}
	}
	wire.WriteTo([]byte(`]}]}`), w, n, err)
}

func (tx *SendTx) GetInputs() []TxInput {
	return copyInputs(tx.Inputs)
}

func (tx *SendTx) String() string {
	return fmt.Sprintf("SendTx{%v -> %v}", tx.Inputs, tx.Outputs)
}

func (tx *SendTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}

func (tx *SendTx) AddInput(st state.AccountGetter, pubkey acm.PublicKey, amt uint64) error {
	addr := pubkey.Address()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return err
	}
	if acc == nil {
		return fmt.Errorf("invalid address %s from pubkey %s", addr, pubkey)
	}
	return tx.AddInputWithSequence(pubkey, amt, acc.Sequence()+1)
}

func (tx *SendTx) AddInputWithSequence(pubkey acm.PublicKey, amt uint64, sequence uint64) error {
	addr := pubkey.Address()
	tx.Inputs = append(tx.Inputs, &TxInput{
		Address:   addr,
		Amount:    amt,
		Sequence:  sequence,
		PublicKey: pubkey,
	})
	return nil
}

func (tx *SendTx) AddOutput(addr acm.Address, amt uint64) error {
	tx.Outputs = append(tx.Outputs, &TxOutput{
		Address: addr,
		Amount:  amt,
	})
	return nil
}

func (tx *SendTx) Sign(chainID string, signingAccounts ...acm.SigningAccount) error {
	if len(signingAccounts) != len(tx.Inputs) {
		return fmt.Errorf("SendTx has %v Inputs but was provided with %v SigningAccounts", len(tx.Inputs),
			len(signingAccounts))
	}
	var err error
	for i, signingAccount := range signingAccounts {
		tx.Inputs[i].PublicKey = signingAccount.PublicKey()
		tx.Inputs[i].Signature, err = acm.ChainSign(signingAccount, chainID, tx)
		if err != nil {
			return fmt.Errorf("could not sign tx %v input %v: %v", tx, tx.Inputs[i], err)
		}
	}
	return nil
}
