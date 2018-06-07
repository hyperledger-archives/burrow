package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/genesis/spec"
)

type GovernanceTx struct {
	Input          *TxInput
	AccountUpdates []spec.TemplateAccount
}

func NewGovTx(st state.AccountGetter, from crypto.Address, accounts ...spec.TemplateAccount) (*GovernanceTx, error) {
	acc, err := st.GetAccount(from)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("could not get account %v", from)
	}

	sequence := acc.Sequence() + 1
	return NewGovTxWithSequence(from, sequence, accounts), nil
}

func NewGovTxWithSequence(from crypto.Address, sequence uint64, accounts []spec.TemplateAccount) *GovernanceTx {
	return &GovernanceTx{
		Input: &TxInput{
			Address:  from,
			Sequence: sequence,
		},
		AccountUpdates: accounts,
	}
}

func (tx *GovernanceTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *GovernanceTx) String() string {
	return fmt.Sprintf("GovernanceTx{%v -> %v}", tx.Input, tx.AccountUpdates)
}
