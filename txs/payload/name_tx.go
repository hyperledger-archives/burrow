package payload

import (
	"fmt"
	"regexp"

	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/names"
)

// Name should be file system lik
// Data should be anything permitted in JSON
var regexpAlphaNum = regexp.MustCompile("^[a-zA-Z0-9._/-@]*$")
var regexpJSON = regexp.MustCompile(`^[a-zA-Z0-9_/ \-+"':,\n\t.{}()\[\]]*$`)

type NameTx struct {
	Input *TxInput
	Name  string
	Data  string
	Fee   uint64
}

func NewNameTx(st state.AccountGetter, from crypto.PublicKey, name, data string, amt, fee uint64) (*NameTx, error) {
	addr := from.Address()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("Invalid address %s from pubkey %s", addr, from)
	}

	sequence := acc.Sequence() + 1
	return NewNameTxWithSequence(from, name, data, amt, fee, sequence), nil
}

func NewNameTxWithSequence(from crypto.PublicKey, name, data string, amt, fee, sequence uint64) *NameTx {
	input := &TxInput{
		Address:  from.Address(),
		Amount:   amt,
		Sequence: sequence,
	}

	return &NameTx{
		Input: input,
		Name:  name,
		Data:  data,
		Fee:   fee,
	}
}

func (tx *NameTx) Type() Type {
	return TypeName
}

func (tx *NameTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *NameTx) ValidateStrings() error {
	if len(tx.Name) == 0 {
		return ErrTxInvalidString{"Name must not be empty"}
	}
	if len(tx.Name) > names.MaxNameLength {
		return ErrTxInvalidString{fmt.Sprintf("Name is too long. Max %d bytes", names.MaxNameLength)}
	}
	if len(tx.Data) > names.MaxDataLength {
		return ErrTxInvalidString{fmt.Sprintf("Data is too long. Max %d bytes", names.MaxDataLength)}
	}

	if !validateNameRegEntryName(tx.Name) {
		return ErrTxInvalidString{fmt.Sprintf("Invalid characters found in NameTx.Name (%s). Only alphanumeric, underscores, dashes, forward slashes, and @ are allowed", tx.Name)}
	}

	if !validateNameRegEntryData(tx.Data) {
		return ErrTxInvalidString{fmt.Sprintf("Invalid characters found in NameTx.Data (%s). Only the kind of things found in a JSON file are allowed", tx.Data)}
	}

	return nil
}

func (tx *NameTx) String() string {
	return fmt.Sprintf("NameTx{%v -> %s: %s}", tx.Input, tx.Name, tx.Data)
}

// filter strings
func validateNameRegEntryName(name string) bool {
	return regexpAlphaNum.Match([]byte(name))
}

func validateNameRegEntryData(data string) bool {
	return regexpJSON.Match([]byte(data))
}
