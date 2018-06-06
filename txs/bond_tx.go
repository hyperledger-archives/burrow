package txs

import (
	"fmt"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/crypto"
)

type BondTx struct {
	PubKey    crypto.PublicKey
	Signature crypto.Signature
	Inputs    []*TxInput
	UnbondTo  []*TxOutput
	txHashMemoizer
}

var _ Tx = &BondTx{}

func NewBondTx(pubkey crypto.PublicKey) (*BondTx, error) {
	return &BondTx{
		PubKey:   pubkey,
		Inputs:   []*TxInput{},
		UnbondTo: []*TxOutput{},
	}, nil
}

func (tx *BondTx) Type() TxType {
	return TxTypeBond
}

func (tx *BondTx) GetInputs() []TxInput {
	return copyInputs(tx.Inputs)
}

func (tx *BondTx) String() string {
	return fmt.Sprintf("BondTx{%v: %v -> %v}", tx.PubKey, tx.Inputs, tx.UnbondTo)
}

func (tx *BondTx) AddInput(st state.AccountGetter, pubkey crypto.PublicKey, amt uint64) error {
	addr := pubkey.Address()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return err
	}
	if acc == nil {
		return fmt.Errorf("Invalid address %s from pubkey %s", addr, pubkey)
	}
	return tx.AddInputWithSequence(pubkey, amt, acc.Sequence()+uint64(1))
}

func (tx *BondTx) AddInputWithSequence(pubkey crypto.PublicKey, amt uint64, sequence uint64) error {
	tx.Inputs = append(tx.Inputs, &TxInput{
		Address:   pubkey.Address(),
		Amount:    amt,
		Sequence:  sequence,
		PublicKey: pubkey,
	})
	return nil
}

func (tx *BondTx) AddOutput(addr crypto.Address, amt uint64) error {
	tx.UnbondTo = append(tx.UnbondTo, &TxOutput{
		Address: addr,
		Amount:  amt,
	})
	return nil
}
