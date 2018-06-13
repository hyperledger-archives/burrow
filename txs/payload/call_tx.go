package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/crypto"
)

type CallTx struct {
	Input *TxInput
	// Pointer since CallTx defines unset 'to' address as inducing account creation
	Address  *crypto.Address
	GasLimit uint64
	Fee      uint64
	// Signing normalisation needs omitempty
	Data []byte `json:",omitempty"`
}

func NewCallTx(st state.AccountGetter, from crypto.PublicKey, to *crypto.Address, data []byte,
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

func NewCallTxWithSequence(from crypto.PublicKey, to *crypto.Address, data []byte,
	amt, gasLimit, fee, sequence uint64) *CallTx {
	input := &TxInput{
		Address:  from.Address(),
		Amount:   amt,
		Sequence: sequence,
	}

	return &CallTx{
		Input:    input,
		Address:  to,
		GasLimit: gasLimit,
		Fee:      fee,
		Data:     data,
	}
}

func (tx *CallTx) Type() Type {
	return TypeCall
}
func (tx *CallTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *CallTx) String() string {
	return fmt.Sprintf("CallTx{%v -> %s: %X}", tx.Input, tx.Address, tx.Data)
}
