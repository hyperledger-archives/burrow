package txs

import (
	"fmt"

	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission/snatives"
)

type PermissionsTx struct {
	Input    *TxInput
	PermArgs snatives.PermArgs
	txHashMemoizer
}

var _ Tx = &PermissionsTx{}

func NewPermissionsTx(st state.AccountGetter, from crypto.PublicKey, args snatives.PermArgs) (*PermissionsTx, error) {
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

func NewPermissionsTxWithSequence(from crypto.PublicKey, args snatives.PermArgs, sequence uint64) *PermissionsTx {
	input := &TxInput{
		Address:   from.Address(),
		Amount:    1, // NOTE: amounts can't be 0 ...
		Sequence:  sequence,
		PublicKey: from,
	}

	return &PermissionsTx{
		Input:    input,
		PermArgs: args,
	}
}

func (tx *PermissionsTx) Type() TxType {
	return TxTypePermissions
}

func (tx *PermissionsTx) GetInputs() []TxInput {
	return []TxInput{*tx.Input}
}

func (tx *PermissionsTx) String() string {
	return fmt.Sprintf("PermissionsTx{%v -> %v}", tx.Input, tx.PermArgs)
}
