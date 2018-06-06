package txs

import (
	"fmt"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/crypto"
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

func (tx *SendTx) GetInputs() []TxInput {
	return copyInputs(tx.Inputs)
}

func (tx *SendTx) Type() TxType {
	return TxTypeSend
}

func (tx *SendTx) String() string {
	return fmt.Sprintf("SendTx{%v -> %v}", tx.Inputs, tx.Outputs)
}

func (tx *SendTx) AddInput(st state.AccountGetter, pubkey crypto.PublicKey, amt uint64) error {
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

func (tx *SendTx) AddInputWithSequence(pubkey crypto.PublicKey, amt uint64, sequence uint64) error {
	addr := pubkey.Address()
	tx.Inputs = append(tx.Inputs, &TxInput{
		Address:   addr,
		Amount:    amt,
		Sequence:  sequence,
		PublicKey: pubkey,
	})
	return nil
}

func (tx *SendTx) AddOutput(addr crypto.Address, amt uint64) error {
	tx.Outputs = append(tx.Outputs, &TxOutput{
		Address: addr,
		Amount:  amt,
	})
	return nil
}

