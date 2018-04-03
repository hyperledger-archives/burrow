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

// TODO: replace with sum-type struct like ResultEvent
type Tx interface {
	WriteSignBytes(chainID string, w io.Writer, n *int, err *error)
	String() string
	GetInputs() []TxInput
	Hash(chainID string) []byte
	Sign(chainID string, signingAccounts ...acm.SigningAccount) error
}

type Encoder interface {
	EncodeTx(tx Tx) ([]byte, error)
}

type Decoder interface {
	DecodeTx(txBytes []byte) (Tx, error)
}

// BroadcastTx or Transact
type Receipt struct {
	TxHash          []byte
	CreatesContract bool
	ContractAddress acm.Address
}

type Wrapper struct {
	Tx `json:"unwrap"`
}

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

// Contract: This function is deterministic and completely reversible.
func jsonEscape(str string) string {
	// TODO: escape without panic
	escapedBytes, err := json.Marshal(str)
	if err != nil {
		panic(fmt.Errorf("error json-escaping string: %s", str))
	}
	return string(escapedBytes)
}
