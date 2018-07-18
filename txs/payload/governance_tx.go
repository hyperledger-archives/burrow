package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/genesis/spec"
)

func NewGovTx(st state.AccountGetter, from crypto.Address, accounts ...*spec.TemplateAccount) (*GovernanceTx, error) {
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

func NewGovTxWithSequence(from crypto.Address, sequence uint64, accounts []*spec.TemplateAccount) *GovernanceTx {
	return &GovernanceTx{
		Inputs: []*TxInput{{
			Address:  from,
			Sequence: sequence,
		}},
		AccountUpdates: accounts,
	}
}

func (tx *GovernanceTx) Type() Type {
	return TypeGovernance
}

func (tx *GovernanceTx) GetInputs() []*TxInput {
	return tx.Inputs
}

func (tx *GovernanceTx) String() string {
	return fmt.Sprintf("GovernanceTx{%v -> %v}", tx.Inputs, tx.AccountUpdates)
}
