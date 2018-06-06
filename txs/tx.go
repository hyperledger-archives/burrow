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

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
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

type TxType int8

// Types of Tx implementations
const (
	TxTypeUnknown = TxType(0x00)
	// Account transactions
	TxTypeSend = TxType(0x01)
	TxTypeCall = TxType(0x02)
	TxTypeName = TxType(0x03)

	// Validation transactions
	TxTypeBond   = TxType(0x11)
	TxTypeUnbond = TxType(0x12)

	// Admin transactions
	TxTypePermissions = TxType(0x21)
	TxTypeGovernance  = TxType(0x22)
)

var txNameFromType = map[TxType]string{
	TxTypeUnknown:     "UnknownTx",
	TxTypeSend:        "SendTx",
	TxTypeCall:        "CallTx",
	TxTypeName:        "NameTx",
	TxTypeBond:        "BondTx",
	TxTypeUnbond:      "UnbondTx",
	TxTypePermissions: "PermissionsTx",
	TxTypeGovernance:  "GovernanceTx",
}

var txTypeFromName = make(map[string]TxType)

func init() {
	for t, n := range txNameFromType {
		txTypeFromName[n] = t
	}
}

//-----------------------------------------------------------------------------

type Tx interface {
	String() string
	GetInputs() []TxInput
	Type() TxType
}

type Encoder interface {
	EncodeTx(tx Tx) ([]byte, error)
}

type Decoder interface {
	DecodeTx(txBytes []byte) (Tx, error)
}

func NewTx(txType TxType) Tx {
	switch txType {
	case TxTypeSend:
		return &SendTx{}
	case TxTypeCall:
		return &CallTx{}
	case TxTypeName:
		return &NameTx{}
	case TxTypeBond:
		return &BondTx{}
	case TxTypeUnbond:
		return &UnbondTx{}
	case TxTypePermissions:
		return &PermissionsTx{}
	}
	return nil
}

func (txType TxType) String() string {
	name, ok := txNameFromType[txType]
	if ok {
		return name
	}
	return "UnknownTx"
}

func TxTypeFromString(name string) TxType {
	return txTypeFromName[name]
}

func (txType TxType) MarshalText() ([]byte, error) {
	return []byte(txType.String()), nil
}

func (txType *TxType) UnmarshalText(data []byte) error {
	*txType = TxTypeFromString(string(data))
	return nil
}

// BroadcastTx or Transact
type Receipt struct {
	TxHash          []byte
	CreatesContract bool
	ContractAddress crypto.Address
}

type Envelope struct {
	Signatures []crypto.Signature
	Body
}

func (env *Envelope) Sign(signingAccounts ...acm.AddressableSigner) error {
	signBytes, err := env.Body.SignBytes()
	if err != nil {
		return err
	}
	for _, sa := range signingAccounts {
		sig, err := sa.Sign(signBytes)
		if err != nil {
			return err
		}
		env.Signatures = append(env.Signatures, sig)
	}
	return nil
}

func SignTx(chainID string, tx Tx, signingAccounts ...acm.AddressableSigner) (*Envelope, error) {
	env := &Envelope{
		Body: Body{
			ChainID: chainID,
			Tx:      tx,
		},
	}
	err := env.Sign(signingAccounts...)
	if err != nil {
		return nil, err
	}
	return env, nil
}

// The
type Body struct {
	ChainID string
	TxType  TxType
	Tx
	txHash []byte
}

// Wrap the Tx in Body required for signing and serialisation
func Wrap(tx Tx) *Body {
	switch t := tx.(type) {
	case Body:
		return &t
	case *Body:
		return t
	}
	return &Body{
		TxType: tx.Type(),
		Tx:     tx,
	}
}

func ChainWrap(chainID string, tx Tx) *Body {
	body := Wrap(tx)
	body.ChainID = chainID
	return body
}

func (body *Body) MustSignBytes() []byte {
	bs, err := body.SignBytes()
	if err != nil {
		panic(err)
	}
	return bs
}

func (body *Body) SignBytes() ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("could not generate canonical SignBytes for tx %v: %v", body.Tx, err)
	}
	return bs, nil
}

// Serialisation intermediate for switching on type
type wrapper struct {
	ChainID string
	TxType  TxType
	Tx      json.RawMessage
}

func (body Body) MarshalJSON() ([]byte, error) {
	bs, err := json.Marshal(body.Tx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(wrapper{
		ChainID: body.ChainID,
		TxType:  body.Type(),
		Tx:      bs,
	})
}

func (body *Body) UnmarshalJSON(data []byte) error {
	w := new(wrapper)
	err := json.Unmarshal(data, w)
	if err != nil {
		return err
	}
	body.ChainID = w.ChainID
	body.TxType = w.TxType
	body.Tx = NewTx(w.TxType)
	return json.Unmarshal(w.Tx, body.Tx)
}

// Get the inner Tx that this Wrapper wraps
func (body *Body) Unwrap() Tx {
	return body.Tx
}

func (body *Body) Hash() []byte {
	if body.txHash == nil {
		return body.Rehash()
	}
	return body.txHash
}

func (body *Body) Rehash() []byte {
	hasher := ripemd160.New()
	hasher.Write(body.MustSignBytes())
	body.txHash = hasher.Sum(nil)
	return body.txHash
}

// Avoid re-hashing the same in-memory Tx
type txHashMemoizer struct {
	txHashBytes []byte
	chainID     string
}

func (body *Body) GenerateReceipt() Receipt {
	receipt := Receipt{
		TxHash: body.Hash(),
	}
	if callTx, ok := body.Tx.(*CallTx); ok {
		receipt.CreatesContract = callTx.Address == nil
		if receipt.CreatesContract {
			receipt.ContractAddress = crypto.NewContractAddress(callTx.Input.Address, callTx.Input.Sequence)
		} else {
			receipt.ContractAddress = *callTx.Address
		}
	}
	return receipt
}

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

//--------------------------------------------------------------------------------

func copyInputs(inputs []*TxInput) []TxInput {
	inputsCopy := make([]TxInput, len(inputs))
	for i, input := range inputs {
		inputsCopy[i] = *input
	}
	return inputsCopy
}
