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

package account

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/tendermint/go-amino"
)

var GlobalPermissionsAddress = crypto.Address(binary.Zero160)

type Addressable interface {
	// Get the 20 byte EVM address of this account
	Address() crypto.Address
	// Public key from which the Address is derived
	PublicKey() crypto.PublicKey
}

// The default immutable interface to an account
type Account interface {
	Addressable
	// The value held by this account in terms of the chain-native token
	Balance() uint64
	// The EVM byte code held by this account (or equivalently, this contract)
	Code() Bytecode
	// The sequence number of this account, incremented each time a mutation of the
	// Account is persisted to the blockchain state
	Sequence() uint64
	// The hash of all the values in this accounts storage (typically the root of some subtree
	// in the merkle tree of global storage state)
	StorageRoot() []byte
	// The permission flags and roles for this account
	Permissions() ptypes.AccountPermissions
	// Obtain a deterministic serialisation of this account
	// (i.e. update order and Go runtime independent)
	Encode() ([]byte, error)
	// String representation of the account
	String() string
}

// MutableAccount structure
type MutableAccount struct {
	concreteAccount *ConcreteAccount
}

type ConcreteAccount struct {
	Address     crypto.Address
	PublicKey   crypto.PublicKey
	Sequence    uint64
	Balance     uint64
	Code        Bytecode
	StorageRoot []byte
	Permissions ptypes.AccountPermissions
}

func NewConcreteAccount(pubKey crypto.PublicKey) *ConcreteAccount {
	return &ConcreteAccount{
		Address:   pubKey.Address(),
		PublicKey: pubKey,
	}
}

func NewConcreteAccountFromSecret(secret string) *ConcreteAccount {
	return NewConcreteAccount(crypto.PrivateKeyFromSecret(secret, crypto.CurveTypeEd25519).GetPublicKey())
}

func (ca ConcreteAccount) Account() Account {
	return ca.MutableAccount()
}

func (ca ConcreteAccount) MutableAccount() *MutableAccount {
	caCopy := ca
	return &MutableAccount{
		concreteAccount: &caCopy,
	}
}

func (ca ConcreteAccount) Encode() ([]byte, error) {
	return cdc.MarshalBinary(ca)
}

func DecodeConcrete(accBytes []byte) (*ConcreteAccount, error) {
	ca := new(ConcreteAccount)
	err := cdc.UnmarshalBinary(accBytes, ca)
	if err != nil {
		return nil, err
	}
	return ca, nil
}

// Conversions
//
// Using the naming convention is this package of 'As<Type>' being
// a conversion from Account to <Type> and 'From<Type>' being conversion
// from <Type> to Account. Conversions are done by copying

// Returns a mutable, serialisable ConcreteAccount by copying from account
func AsConcreteAccount(account Account) *ConcreteAccount {
	if account == nil {
		return nil
	}
	return &ConcreteAccount{
		Address:     account.Address(),
		PublicKey:   account.PublicKey(),
		Balance:     account.Balance(),
		Code:        account.Code(),
		Sequence:    account.Sequence(),
		StorageRoot: account.StorageRoot(),
		Permissions: account.Permissions(),
	}
}

// Creates an otherwise zeroed Account from an Addressable and returns it as MutableAccount
func FromAddressable(addressable Addressable) *MutableAccount {
	ca := &ConcreteAccount{
		Address:   addressable.Address(),
		PublicKey: addressable.PublicKey(),
		// Since nil slices and maps compare differently to empty ones
		Code:        Bytecode{},
		StorageRoot: []byte{},
		Permissions: ptypes.AccountPermissions{
			Roles: []string{},
		},
	}
	return ca.MutableAccount()
}

// Returns a MutableAccount by copying from account
func AsMutableAccount(account Account) *MutableAccount {
	if account == nil {
		return nil
	}
	return AsConcreteAccount(account).MutableAccount()
}

///---- Getter methods
func (acc MutableAccount) Address() crypto.Address     { return acc.concreteAccount.Address }
func (acc MutableAccount) PublicKey() crypto.PublicKey { return acc.concreteAccount.PublicKey }
func (acc MutableAccount) Balance() uint64             { return acc.concreteAccount.Balance }
func (acc MutableAccount) Code() Bytecode              { return acc.concreteAccount.Code }
func (acc MutableAccount) Sequence() uint64            { return acc.concreteAccount.Sequence }
func (acc MutableAccount) StorageRoot() []byte         { return acc.concreteAccount.StorageRoot }
func (acc MutableAccount) Permissions() ptypes.AccountPermissions {
	return acc.concreteAccount.Permissions
}

///---- Mutable methods
// Set public key (needed for lazy initialisation), should also set the dependent address
func (acc *MutableAccount) SetPublicKey(publicKey crypto.PublicKey) {
	acc.concreteAccount.PublicKey = publicKey
}

func (acc *MutableAccount) SubtractFromBalance(amount uint64) error {
	if amount > acc.Balance() {
		return fmt.Errorf("insufficient funds: attempt to subtract %v from the balance of %s",
			amount, acc.Address())
	}
	acc.concreteAccount.Balance -= amount
	return nil
}

func (acc *MutableAccount) AddToBalance(amount uint64) error {
	if binary.IsUint64SumOverflow(acc.Balance(), amount) {
		return fmt.Errorf("uint64 overflow: attempt to add %v to the balance of %s",
			amount, acc.Address())
	}
	acc.concreteAccount.Balance += amount
	return nil
}

func (acc *MutableAccount) SetCode(code []byte) error {
	acc.concreteAccount.Code = code
	return nil
}

func (acc *MutableAccount) IncSequence() {
	acc.concreteAccount.Sequence++
}

func (acc *MutableAccount) SetStorageRoot(storageRoot []byte) error {
	acc.concreteAccount.StorageRoot = storageRoot
	return nil
}

func (acc *MutableAccount) SetPermissions(permissions ptypes.AccountPermissions) error {
	acc.concreteAccount.Permissions = permissions
	return nil
}

func (acc *MutableAccount) MutablePermissions() *ptypes.AccountPermissions {
	return &acc.concreteAccount.Permissions
}

///---- Serialisation methods

var cdc = amino.NewCodec()

func (acc MutableAccount) Encode() ([]byte, error) {
	return acc.concreteAccount.Encode()
}

func Decode(accBytes []byte) (*MutableAccount, error) {
	ca, err := DecodeConcrete(accBytes)
	if err != nil {
		return nil, err
	}
	return &MutableAccount{
		concreteAccount: ca,
	}, nil
}

func (acc MutableAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(acc.concreteAccount)
}
func (acc *MutableAccount) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, &acc.concreteAccount)
	if err != nil {
		return err
	}
	return nil
}

func (acc MutableAccount) String() string {
	return fmt.Sprintf("Account{Address: %s; Sequence: %v; PublicKey: %v Balance: %v; CodeBytes: %v; StorageRoot: 0x%X; Permissions: %s}",
		acc.Address(), acc.Sequence(), acc.PublicKey(), acc.Balance(), len(acc.Code()), acc.StorageRoot(), acc.Permissions())
}
