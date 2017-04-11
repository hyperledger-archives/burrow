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
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	ptypes "github.com/hyperledger/burrow/permission/types"

	"github.com/tendermint/go-crypto"
)

type AccountGetter interface {
	GetAccount(addr []byte) *acm.Account
}

//----------------------------------------------------------------------------
// SendTx interface for adding inputs/outputs and adding signatures

func NewSendTx() *SendTx {
	return &SendTx{
		Inputs:  []*TxInput{},
		Outputs: []*TxOutput{},
	}
}

func (tx *SendTx) AddInput(st AccountGetter, pubkey crypto.PubKey, amt int64) error {
	addr := pubkey.Address()
	acc := st.GetAccount(addr)
	if acc == nil {
		return fmt.Errorf("Invalid address %X from pubkey %X", addr, pubkey)
	}
	return tx.AddInputWithNonce(pubkey, amt, acc.Sequence+1)
}

func (tx *SendTx) AddInputWithNonce(pubkey crypto.PubKey, amt int64, nonce int) error {
	addr := pubkey.Address()
	tx.Inputs = append(tx.Inputs, &TxInput{
		Address:   addr,
		Amount:    amt,
		Sequence:  nonce,
		Signature: crypto.SignatureEd25519{},
		PubKey:    pubkey,
	})
	return nil
}

func (tx *SendTx) AddOutput(addr []byte, amt int64) error {
	tx.Outputs = append(tx.Outputs, &TxOutput{
		Address: addr,
		Amount:  amt,
	})
	return nil
}

func (tx *SendTx) SignInput(chainID string, i int, privAccount *acm.PrivAccount) error {
	if i >= len(tx.Inputs) {
		return fmt.Errorf("Index %v is greater than number of inputs (%v)", i, len(tx.Inputs))
	}
	tx.Inputs[i].PubKey = privAccount.PubKey
	tx.Inputs[i].Signature = privAccount.Sign(chainID, tx)
	return nil
}

//----------------------------------------------------------------------------
// CallTx interface for creating tx

func NewCallTx(st AccountGetter, from crypto.PubKey, to, data []byte, amt, gasLimit, fee int64) (*CallTx, error) {
	addr := from.Address()
	acc := st.GetAccount(addr)
	if acc == nil {
		return nil, fmt.Errorf("Invalid address %X from pubkey %X", addr, from)
	}

	nonce := acc.Sequence + 1
	return NewCallTxWithNonce(from, to, data, amt, gasLimit, fee, nonce), nil
}

func NewCallTxWithNonce(from crypto.PubKey, to, data []byte, amt, gasLimit, fee int64, nonce int) *CallTx {
	addr := from.Address()
	input := &TxInput{
		Address:   addr,
		Amount:    amt,
		Sequence:  nonce,
		Signature: crypto.SignatureEd25519{},
		PubKey:    from,
	}

	return &CallTx{
		Input:    input,
		Address:  to,
		GasLimit: gasLimit,
		Fee:      fee,
		Data:     data,
	}
}

func (tx *CallTx) Sign(chainID string, privAccount *acm.PrivAccount) {
	tx.Input.PubKey = privAccount.PubKey
	tx.Input.Signature = privAccount.Sign(chainID, tx)
}

//----------------------------------------------------------------------------
// NameTx interface for creating tx

func NewNameTx(st AccountGetter, from crypto.PubKey, name, data string, amt, fee int64) (*NameTx, error) {
	addr := from.Address()
	acc := st.GetAccount(addr)
	if acc == nil {
		return nil, fmt.Errorf("Invalid address %X from pubkey %X", addr, from)
	}

	nonce := acc.Sequence + 1
	return NewNameTxWithNonce(from, name, data, amt, fee, nonce), nil
}

func NewNameTxWithNonce(from crypto.PubKey, name, data string, amt, fee int64, nonce int) *NameTx {
	addr := from.Address()
	input := &TxInput{
		Address:   addr,
		Amount:    amt,
		Sequence:  nonce,
		Signature: crypto.SignatureEd25519{},
		PubKey:    from,
	}

	return &NameTx{
		Input: input,
		Name:  name,
		Data:  data,
		Fee:   fee,
	}
}

func (tx *NameTx) Sign(chainID string, privAccount *acm.PrivAccount) {
	tx.Input.PubKey = privAccount.PubKey
	tx.Input.Signature = privAccount.Sign(chainID, tx)
}

//----------------------------------------------------------------------------
// BondTx interface for adding inputs/outputs and adding signatures

func NewBondTx(pubkey crypto.PubKey) (*BondTx, error) {
	pubkeyEd, ok := pubkey.(crypto.PubKeyEd25519)
	if !ok {
		return nil, fmt.Errorf("Pubkey must be ed25519")
	}
	return &BondTx{
		PubKey:   pubkeyEd,
		Inputs:   []*TxInput{},
		UnbondTo: []*TxOutput{},
	}, nil
}

func (tx *BondTx) AddInput(st AccountGetter, pubkey crypto.PubKey, amt int64) error {
	addr := pubkey.Address()
	acc := st.GetAccount(addr)
	if acc == nil {
		return fmt.Errorf("Invalid address %X from pubkey %X", addr, pubkey)
	}
	return tx.AddInputWithNonce(pubkey, amt, acc.Sequence+1)
}

func (tx *BondTx) AddInputWithNonce(pubkey crypto.PubKey, amt int64, nonce int) error {
	addr := pubkey.Address()
	tx.Inputs = append(tx.Inputs, &TxInput{
		Address:   addr,
		Amount:    amt,
		Sequence:  nonce,
		Signature: crypto.SignatureEd25519{},
		PubKey:    pubkey,
	})
	return nil
}

func (tx *BondTx) AddOutput(addr []byte, amt int64) error {
	tx.UnbondTo = append(tx.UnbondTo, &TxOutput{
		Address: addr,
		Amount:  amt,
	})
	return nil
}

func (tx *BondTx) SignBond(chainID string, privAccount *acm.PrivAccount) error {
	sig := privAccount.Sign(chainID, tx)
	sigEd, ok := sig.(crypto.SignatureEd25519)
	if !ok {
		return fmt.Errorf("Bond signer must be ED25519")
	}
	tx.Signature = sigEd
	return nil
}

func (tx *BondTx) SignInput(chainID string, i int, privAccount *acm.PrivAccount) error {
	if i >= len(tx.Inputs) {
		return fmt.Errorf("Index %v is greater than number of inputs (%v)", i, len(tx.Inputs))
	}
	tx.Inputs[i].PubKey = privAccount.PubKey
	tx.Inputs[i].Signature = privAccount.Sign(chainID, tx)
	return nil
}

//----------------------------------------------------------------------
// UnbondTx interface for creating tx

func NewUnbondTx(addr []byte, height int) *UnbondTx {
	return &UnbondTx{
		Address: addr,
		Height:  height,
	}
}

func (tx *UnbondTx) Sign(chainID string, privAccount *acm.PrivAccount) {
	tx.Signature = privAccount.Sign(chainID, tx).(crypto.SignatureEd25519)
}

//----------------------------------------------------------------------
// RebondTx interface for creating tx

func NewRebondTx(addr []byte, height int) *RebondTx {
	return &RebondTx{
		Address: addr,
		Height:  height,
	}
}

func (tx *RebondTx) Sign(chainID string, privAccount *acm.PrivAccount) {
	tx.Signature = privAccount.Sign(chainID, tx).(crypto.SignatureEd25519)
}

//----------------------------------------------------------------------------
// PermissionsTx interface for creating tx

func NewPermissionsTx(st AccountGetter, from crypto.PubKey, args ptypes.PermArgs) (*PermissionsTx, error) {
	addr := from.Address()
	acc := st.GetAccount(addr)
	if acc == nil {
		return nil, fmt.Errorf("Invalid address %X from pubkey %X", addr, from)
	}

	nonce := acc.Sequence + 1
	return NewPermissionsTxWithNonce(from, args, nonce), nil
}

func NewPermissionsTxWithNonce(from crypto.PubKey, args ptypes.PermArgs, nonce int) *PermissionsTx {
	addr := from.Address()
	input := &TxInput{
		Address:   addr,
		Amount:    1, // NOTE: amounts can't be 0 ...
		Sequence:  nonce,
		Signature: crypto.SignatureEd25519{},
		PubKey:    from,
	}

	return &PermissionsTx{
		Input:    input,
		PermArgs: args,
	}
}

func (tx *PermissionsTx) Sign(chainID string, privAccount *acm.PrivAccount) {
	tx.Input.PubKey = privAccount.PubKey
	tx.Input.Signature = privAccount.Sign(chainID, tx)
}
