// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package txs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	ptypes "github.com/hyperledger/burrow/permission"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	"golang.org/x/crypto/ripemd160"
)

var (
	ErrTxInvalidAddress    = errors.New("error invalid address")
	ErrTxDuplicateAddress  = errors.New("error duplicate address")
	ErrTxInvalidAmount     = errors.New("error invalid amount")
	ErrTxInsufficientFunds = errors.New("error insufficient funds")
	ErrTxUnknownPubKey     = errors.New("error unknown pubkey")
	ErrTxInvalidPubKey     = errors.New("error invalid pubkey")
	ErrTxInvalidSignature  = errors.New("error invalid signature")
)

type ErrTxInvalidString struct {
	Msg string
}

func (e ErrTxInvalidString) Error() string {
	return e.Msg
}

type ErrTxInvalidSequence struct {
	Got      uint64
	Expected uint64
}

func (e ErrTxInvalidSequence) Error() string {
	return fmt.Sprintf("Error invalid sequence. Got %d, expected %d", e.Got, e.Expected)
}

/*
Tx (Transaction) is an atomic operation on the ledger state.

Account Txs:
 - SendTx         Send coins to address
 - CallTx         Send a msg to a contract that runs in the vm
 - NameTx	  Store some value under a name in the global namereg

Validation Txs:
 - BondTx         New validator posts a bond
 - UnbondTx       Validator leaves

Admin Txs:
 - PermissionsTx
*/

// Types of Tx implementations
const (
	// Account transactions
	TxTypeSend = byte(0x01)
	TxTypeCall = byte(0x02)
	TxTypeName = byte(0x03)

	// Validation transactions
	TxTypeBond   = byte(0x11)
	TxTypeUnbond = byte(0x12)
	TxTypeRebond = byte(0x13)

	// Admin transactions
	TxTypePermissions = byte(0x1f)
)

var mapper = data.NewMapper(Wrapper{}).
	RegisterImplementation(&SendTx{}, "send_tx", TxTypeSend).
	RegisterImplementation(&CallTx{}, "call_tx", TxTypeCall).
	RegisterImplementation(&NameTx{}, "name_tx", TxTypeName).
	RegisterImplementation(&BondTx{}, "bond_tx", TxTypeBond).
	RegisterImplementation(&UnbondTx{}, "unbond_tx", TxTypeUnbond).
	RegisterImplementation(&RebondTx{}, "rebond_tx", TxTypeRebond).
	RegisterImplementation(&PermissionsTx{}, "permissions_tx", TxTypePermissions)

//-----------------------------------------------------------------------------

type (
	// TODO: replace with sum-type struct like ResultEvent
	Tx interface {
		WriteSignBytes(chainID string, w io.Writer, n *int, err *error)
		String() string
		GetInputs() []TxInput
		Hash(chainID string) []byte
	}

	Wrapper struct {
		Tx `json:"unwrap"`
	}

	Encoder interface {
		EncodeTx(tx Tx) ([]byte, error)
	}

	Decoder interface {
		DecodeTx(txBytes []byte) (Tx, error)
	}

	TxInput struct {
		Address   acm.Address
		Amount    uint64
		Sequence  uint64
		Signature acm.Signature
		PublicKey acm.PublicKey
	}

	TxOutput struct {
		Address acm.Address
		Amount  uint64
	}

	// BroadcastTx or Transact
	Receipt struct {
		TxHash          []byte
		CreatesContract bool
		ContractAddress acm.Address
	}

	//-------------------
	// Transaction Types
	SendTx struct {
		Inputs  []*TxInput
		Outputs []*TxOutput
		txHashMemoizer
	}

	NameTx struct {
		Input *TxInput
		Name  string
		Data  string
		Fee   uint64
		txHashMemoizer
	}

	CallTx struct {
		Input *TxInput
		// Pointer since CallTx defines unset 'to' address as inducing account creation
		Address  *acm.Address
		GasLimit uint64
		Fee      uint64
		Data     []byte
		txHashMemoizer
	}

	PermissionsTx struct {
		Input    *TxInput
		PermArgs ptypes.PermArgs
		txHashMemoizer
	}

	// Out of service
	BondTx struct {
		PubKey    acm.PublicKey
		Signature acm.Signature
		Inputs    []*TxInput
		UnbondTo  []*TxOutput
		txHashMemoizer
	}

	UnbondTx struct {
		Address   acm.Address
		Height    int
		Signature acm.Signature
		txHashMemoizer
	}

	RebondTx struct {
		Address   acm.Address
		Height    int
		Signature acm.Signature
		txHashMemoizer
	}
)

// Wrap the Tx in a struct that allows for go-wire JSON serialisation
func Wrap(tx Tx) Wrapper {
	if txWrapped, ok := tx.(Wrapper); ok {
		return txWrapped
	}
	return Wrapper{
		Tx: tx,
	}
}

// A serialisation wrapper that is itself a Tx
func (txw Wrapper) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	txw.Tx.WriteSignBytes(chainID, w, n, err)
}

func (txw Wrapper) MarshalJSON() ([]byte, error) {
	return mapper.ToJSON(txw.Tx)
}

func (txw *Wrapper) UnmarshalJSON(data []byte) (err error) {
	parsed, err := mapper.FromJSON(data)
	if err == nil && parsed != nil {
		txw.Tx = parsed.(Tx)
	}
	return err
}

// Get the inner Tx that this Wrapper wraps
func (txw *Wrapper) Unwrap() Tx {
	return txw.Tx
}

func (txIn *TxInput) ValidateBasic() error {
	if len(txIn.Address) != 20 {
		return ErrTxInvalidAddress
	}
	if txIn.Amount == 0 {
		return ErrTxInvalidAmount
	}
	return nil
}

func (txIn *TxInput) WriteSignBytes(w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"address":"%s","amount":%v,"sequence":%v}`, txIn.Address, txIn.Amount, txIn.Sequence)), w, n, err)
}

func (txIn *TxInput) String() string {
	return fmt.Sprintf("TxInput{%s,%v,%v,%v,%v}", txIn.Address, txIn.Amount, txIn.Sequence, txIn.Signature, txIn.PublicKey)
}

//-----------------------------------------------------------------------------

func (txOut *TxOutput) ValidateBasic() error {
	if len(txOut.Address) != 20 {
		return ErrTxInvalidAddress
	}
	if txOut.Amount == 0 {
		return ErrTxInvalidAmount
	}
	return nil
}

func (txOut *TxOutput) WriteSignBytes(w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"address":"%s","amount":%v}`, txOut.Address, txOut.Amount)), w, n, err)
}

func (txOut *TxOutput) String() string {
	return fmt.Sprintf("TxOutput{%s,%v}", txOut.Address, txOut.Amount)
}

//-----------------------------------------------------------------------------

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

//-----------------------------------------------------------------------------

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

//-----------------------------------------------------------------------------

func (tx *NameTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s`, jsonEscape(chainID))), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"tx":[%v,{"data":%s,"fee":%v`, TxTypeName, jsonEscape(tx.Data), tx.Fee)), w, n, err)
	wire.WriteTo([]byte(`,"input":`), w, n, err)
	tx.Input.WriteSignBytes(w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"name":%s`, jsonEscape(tx.Name))), w, n, err)
	wire.WriteTo([]byte(`}]}`), w, n, err)
}

func (tx *NameTx) GetInputs() []TxInput {
	return []TxInput{*tx.Input}
}

func (tx *NameTx) ValidateStrings() error {
	if len(tx.Name) == 0 {
		return ErrTxInvalidString{"Name must not be empty"}
	}
	if len(tx.Name) > MaxNameLength {
		return ErrTxInvalidString{fmt.Sprintf("Name is too long. Max %d bytes", MaxNameLength)}
	}
	if len(tx.Data) > MaxDataLength {
		return ErrTxInvalidString{fmt.Sprintf("Data is too long. Max %d bytes", MaxDataLength)}
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

func (tx *NameTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}

//-----------------------------------------------------------------------------

func (tx *BondTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s`, jsonEscape(chainID))), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"tx":[%v,{"inputs":[`, TxTypeBond)), w, n, err)
	for i, in := range tx.Inputs {
		in.WriteSignBytes(w, n, err)
		if i != len(tx.Inputs)-1 {
			wire.WriteTo([]byte(","), w, n, err)
		}
	}
	wire.WriteTo([]byte(fmt.Sprintf(`],"pub_key":`)), w, n, err)
	wire.WriteTo(wire.JSONBytes(tx.PubKey), w, n, err)
	wire.WriteTo([]byte(`,"unbond_to":[`), w, n, err)
	for i, out := range tx.UnbondTo {
		out.WriteSignBytes(w, n, err)
		if i != len(tx.UnbondTo)-1 {
			wire.WriteTo([]byte(","), w, n, err)
		}
	}
	wire.WriteTo([]byte(`]}]}`), w, n, err)
}

func (tx *BondTx) GetInputs() []TxInput {
	return copyInputs(tx.Inputs)
}

func (tx *BondTx) String() string {
	return fmt.Sprintf("BondTx{%v: %v -> %v}", tx.PubKey, tx.Inputs, tx.UnbondTo)
}

func (tx *BondTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}

//-----------------------------------------------------------------------------

func (tx *UnbondTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s`, jsonEscape(chainID))), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"tx":[%v,{"address":"%s","height":%v}]}`, TxTypeUnbond, tx.Address, tx.Height)), w, n, err)
}

func (tx *UnbondTx) GetInputs() []TxInput {
	return nil
}

func (tx *UnbondTx) String() string {
	return fmt.Sprintf("UnbondTx{%s,%v,%v}", tx.Address, tx.Height, tx.Signature)
}

func (tx *UnbondTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}

//-----------------------------------------------------------------------------

func (tx *RebondTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s`, jsonEscape(chainID))), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"tx":[%v,{"address":"%s","height":%v}]}`, TxTypeRebond, tx.Address, tx.Height)), w, n, err)
}

func (tx *RebondTx) GetInputs() []TxInput {
	return nil
}

func (tx *RebondTx) String() string {
	return fmt.Sprintf("RebondTx{%s,%v,%v}", tx.Address, tx.Height, tx.Signature)
}

func (tx *RebondTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}

//-----------------------------------------------------------------------------

func (tx *PermissionsTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s`, jsonEscape(chainID))), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"tx":[%v,{"args":"`, TxTypePermissions)), w, n, err)
	wire.WriteJSON(&tx.PermArgs, w, n, err)
	wire.WriteTo([]byte(`","input":`), w, n, err)
	tx.Input.WriteSignBytes(w, n, err)
	wire.WriteTo([]byte(`}]}`), w, n, err)
}

func (tx *PermissionsTx) GetInputs() []TxInput {
	return []TxInput{*tx.Input}
}

func (tx *PermissionsTx) String() string {
	return fmt.Sprintf("PermissionsTx{%v -> %v}", tx.Input, tx.PermArgs)
}

func (tx *PermissionsTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}

//-----------------------------------------------------------------------------

// Avoid re-hashing the same in-memory Tx
type txHashMemoizer struct {
	txHashBytes []byte
	chainID     string
}

func (thm *txHashMemoizer) hash(chainID string, tx Tx) []byte {
	if thm.txHashBytes == nil || thm.chainID != chainID {
		thm.chainID = chainID
		thm.txHashBytes = TxHash(chainID, tx)
	}
	return thm.txHashBytes
}

func TxHash(chainID string, tx Tx) []byte {
	signBytes := acm.SignBytes(chainID, tx)
	hasher := ripemd160.New()
	hasher.Write(signBytes)
	// Calling Sum(nil) just gives us the digest with nothing prefixed
	return hasher.Sum(nil)
}

func GenerateReceipt(chainId string, tx Tx) Receipt {
	receipt := Receipt{
		TxHash: tx.Hash(chainId),
	}
	if callTx, ok := tx.(*CallTx); ok {
		receipt.CreatesContract = callTx.Address == nil
		if receipt.CreatesContract {
			receipt.ContractAddress = acm.NewContractAddress(callTx.Input.Address, callTx.Input.Sequence)
		} else {
			receipt.ContractAddress = *callTx.Address
		}
	}
	return receipt
}

//--------------------------------------------------------------------------------

func copyInputs(inputs []*TxInput) []TxInput {
	inputsCopy := make([]TxInput, len(inputs))
	for i, input := range inputs {
		inputsCopy[i] = *input
	}
	return inputsCopy
}

// Contract: This function is deterministic and completely reversible.
func jsonEscape(str string) string {
	// TODO: escape without panic
	escapedBytes, err := json.Marshal(str)
	if err != nil {
		panic(fmt.Errorf("error json-escaping string: %s", str))
	}
	return string(escapedBytes)
}
