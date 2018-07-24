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

package acm

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/permission"
	"github.com/tendermint/go-amino"
)

var GlobalPermissionsAddress = crypto.Address(binary.Zero160)

// The default immutable interface to an account
type Account interface {
	crypto.Addressable
	// The value held by this account in terms of the chain-native token
	Balance() uint64
	// The EVM byte code held by this account (or equivalently, this contract)
	Code() Bytecode
	// The sequence number of this account, incremented each time a mutation of the
	// Account is persisted to the blockchain state
	Sequence() uint64
	// The permission flags and roles for this account
	Permissions() permission.AccountPermissions
	// Obtain a deterministic serialisation of this account
	// (i.e. update order and Go runtime independent)
	Encode() ([]byte, error)
	// String representation of the account
	String() string
	// Get tags for this account
	Tagged() query.Tagged
}

// MutableAccount structure
type MutableAccount struct {
	concreteAccount *ConcreteAccount
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

// Wrap a copy of ConcreteAccount in a MutableAccount
func (ca ConcreteAccount) MutableAccount() *MutableAccount {
	return &MutableAccount{
		concreteAccount: &ca,
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
		Permissions: account.Permissions(),
	}
}

// Creates an otherwise zeroed Account from an Addressable and returns it as MutableAccount
func FromAddressable(addressable crypto.Addressable) *MutableAccount {
	ca := &ConcreteAccount{
		Address:   addressable.Address(),
		PublicKey: addressable.PublicKey(),
		// Since nil slices and maps compare differently to empty ones
		Code: Bytecode{},
		Permissions: permission.AccountPermissions{
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

var _ Account = &MutableAccount{}

func (acc ConcreteAccount) String() string {
	return fmt.Sprintf("ConcreteAccount{Address: %s; Sequence: %v; PublicKey: %v Balance: %v; CodeLength: %v; Permissions: %s}",
		acc.Address, acc.Sequence, acc.PublicKey, acc.Balance, len(acc.Code), acc.Permissions)
}

///---- Getter methods
func (acc MutableAccount) Address() crypto.Address     { return acc.concreteAccount.Address }
func (acc MutableAccount) PublicKey() crypto.PublicKey { return acc.concreteAccount.PublicKey }
func (acc MutableAccount) Balance() uint64             { return acc.concreteAccount.Balance }
func (acc MutableAccount) Code() Bytecode              { return acc.concreteAccount.Code }
func (acc MutableAccount) Sequence() uint64            { return acc.concreteAccount.Sequence }
func (acc MutableAccount) Permissions() permission.AccountPermissions {
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

func (acc *MutableAccount) SetBalance(amount uint64) error {
	acc.concreteAccount.Balance = amount
	return nil
}

func (acc *MutableAccount) SetCode(code []byte) error {
	acc.concreteAccount.Code = code
	return nil
}

func (acc *MutableAccount) IncSequence() {
	acc.concreteAccount.Sequence++
}

func (acc *MutableAccount) SetPermissions(accPerms permission.AccountPermissions) error {
	if !accPerms.Base.Perms.IsValid() {
		return fmt.Errorf("attempt to set invalid perm 0%b on account %v", accPerms.Base.Perms, acc)
	}
	acc.concreteAccount.Permissions = accPerms
	return nil
}

func (acc *MutableAccount) MutablePermissions() *permission.AccountPermissions {
	return &acc.concreteAccount.Permissions
}

type TaggedAccount struct {
	*MutableAccount
	query.Tagged
}

func (acc *MutableAccount) Tagged() query.Tagged {
	return &TaggedAccount{
		MutableAccount: acc,
		Tagged:         query.MustReflectTags(acc),
	}
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
	return fmt.Sprintf("MutableAccount{%s}", acc.concreteAccount.String())
}
