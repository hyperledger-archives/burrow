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
	ptypes "github.com/hyperledger/burrow/permission"
)

//----------------------------------------------------------------------------
// SendTx interface for adding inputs/outputs and adding signatures

func NewSendTx() *SendTx {
	return &SendTx{
		Inputs:  []*TxInput{},
		Outputs: []*TxOutput{},
	}
}

func (tx *SendTx) AddInput(st acm.Getter, pubkey acm.PublicKey, amt uint64) error {
	addr := pubkey.Address()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return err
	}
	if acc == nil {
		return fmt.Errorf("invalid address %s from pubkey %s", addr, pubkey)
	}
	return tx.AddInputWithSequence(pubkey, amt, acc.Sequence()+1)
}

func (tx *SendTx) AddInputWithSequence(pubkey acm.PublicKey, amt uint64, sequence uint64) error {
	addr := pubkey.Address()
	tx.Inputs = append(tx.Inputs, &TxInput{
		Address:  addr,
		Amount:   amt,
		Sequence: sequence,
		PubKey:   pubkey,
	})
	return nil
}

func (tx *SendTx) AddOutput(addr acm.Address, amt uint64) error {
	tx.Outputs = append(tx.Outputs, &TxOutput{
		Address: addr,
		Amount:  amt,
	})
	return nil
}

func (tx *SendTx) SignInput(chainID string, i int, privAccount acm.PrivateAccount) error {
	if i >= len(tx.Inputs) {
		return fmt.Errorf("Index %v is greater than number of inputs (%v)", i, len(tx.Inputs))
	}
	tx.Inputs[i].PubKey = privAccount.PublicKey()
	tx.Inputs[i].Signature = acm.ChainSign(privAccount, chainID, tx)
	return nil
}

//----------------------------------------------------------------------------
// CallTx interface for creating tx

func NewCallTx(st acm.Getter, from acm.PublicKey, to *acm.Address, data []byte,
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

func NewCallTxWithSequence(from acm.PublicKey, to *acm.Address, data []byte,
	amt, gasLimit, fee, sequence uint64) *CallTx {
	input := &TxInput{
		Address:  from.Address(),
		Amount:   amt,
		Sequence: sequence,
		PubKey:   from,
	}

	return &CallTx{
		Input:    input,
		Address:  to,
		GasLimit: gasLimit,
		Fee:      fee,
		Data:     data,
	}
}

func (tx *CallTx) Sign(chainID string, privAccount acm.PrivateAccount) {
	tx.Input.PubKey = privAccount.PublicKey()
	tx.Input.Signature = acm.ChainSign(privAccount, chainID, tx)
}

//----------------------------------------------------------------------------
// NameTx interface for creating tx

func NewNameTx(st acm.Getter, from acm.PublicKey, name, data string, amt, fee uint64) (*NameTx, error) {
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

func NewNameTxWithSequence(from acm.PublicKey, name, data string, amt, fee, sequence uint64) *NameTx {
	input := &TxInput{
		Address:  from.Address(),
		Amount:   amt,
		Sequence: sequence,
		PubKey:   from,
	}

	return &NameTx{
		Input: input,
		Name:  name,
		Data:  data,
		Fee:   fee,
	}
}

func (tx *NameTx) Sign(chainID string, privAccount acm.PrivateAccount) {
	tx.Input.PubKey = privAccount.PublicKey()
	tx.Input.Signature = acm.ChainSign(privAccount, chainID, tx)
}

//----------------------------------------------------------------------------
// BondTx interface for adding inputs/outputs and adding signatures

func NewBondTx(pubkey acm.PublicKey) (*BondTx, error) {
	return &BondTx{
		PubKey:   pubkey,
		Inputs:   []*TxInput{},
		UnbondTo: []*TxOutput{},
	}, nil
}

func (tx *BondTx) AddInput(st acm.Getter, pubkey acm.PublicKey, amt uint64) error {
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

func (tx *BondTx) AddInputWithSequence(pubkey acm.PublicKey, amt uint64, sequence uint64) error {
	tx.Inputs = append(tx.Inputs, &TxInput{
		Address:  pubkey.Address(),
		Amount:   amt,
		Sequence: sequence,
		PubKey:   pubkey,
	})
	return nil
}

func (tx *BondTx) AddOutput(addr acm.Address, amt uint64) error {
	tx.UnbondTo = append(tx.UnbondTo, &TxOutput{
		Address: addr,
		Amount:  amt,
	})
	return nil
}

func (tx *BondTx) SignBond(chainID string, privAccount acm.PrivateAccount) error {
	tx.Signature = acm.ChainSign(privAccount, chainID, tx)
	return nil
}

func (tx *BondTx) SignInput(chainID string, i int, privAccount acm.PrivateAccount) error {
	if i >= len(tx.Inputs) {
		return fmt.Errorf("Index %v is greater than number of inputs (%v)", i, len(tx.Inputs))
	}
	tx.Inputs[i].PubKey = privAccount.PublicKey()
	tx.Inputs[i].Signature = acm.ChainSign(privAccount, chainID, tx)
	return nil
}

//----------------------------------------------------------------------
// UnbondTx interface for creating tx

func NewUnbondTx(addr acm.Address, height int) *UnbondTx {
	return &UnbondTx{
		Address: addr,
		Height:  height,
	}
}

func (tx *UnbondTx) Sign(chainID string, privAccount acm.PrivateAccount) {
	tx.Signature = acm.ChainSign(privAccount, chainID, tx)
}

//----------------------------------------------------------------------
// RebondTx interface for creating tx

func NewRebondTx(addr acm.Address, height int) *RebondTx {
	return &RebondTx{
		Address: addr,
		Height:  height,
	}
}

func (tx *RebondTx) Sign(chainID string, privAccount acm.PrivateAccount) {
	tx.Signature = acm.ChainSign(privAccount, chainID, tx)
}

//----------------------------------------------------------------------------
// PermissionsTx interface for creating tx

func NewPermissionsTx(st acm.Getter, from acm.PublicKey, args *ptypes.PermArgs) (*PermissionsTx, error) {
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

func NewPermissionsTxWithSequence(from acm.PublicKey, args *ptypes.PermArgs, sequence uint64) *PermissionsTx {
	input := &TxInput{
		Address:  from.Address(),
		Amount:   1, // NOTE: amounts can't be 0 ...
		Sequence: sequence,
		PubKey:   from,
	}

	return &PermissionsTx{
		Input:    input,
		PermArgs: args,
	}
}

func (tx *PermissionsTx) Sign(chainID string, privAccount acm.PrivateAccount) {
	tx.Input.PubKey = privAccount.PublicKey()
	tx.Input.Signature = acm.ChainSign(privAccount, chainID, tx)
}
