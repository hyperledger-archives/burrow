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
	"bytes"
	"fmt"

	amino "github.com/tendermint/go-amino"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/permission"
)

var GlobalPermissionsAddress = crypto.Address(binary.Zero160)

func NewAccount(pubKey crypto.PublicKey) *Account {
	return &Account{
		Address:   pubKey.GetAddress(),
		PublicKey: pubKey,
	}
}

func NewAccountFromSecret(secret string) *Account {
	return NewAccount(crypto.PrivateKeyFromSecret(secret, crypto.CurveTypeEd25519).GetPublicKey())
}

func (acc *Account) GetAddress() crypto.Address {
	return acc.Address
}

///---- Serialisation methods

var cdc = amino.NewCodec()

func (acc *Account) Encode() ([]byte, error) {
	return cdc.MarshalBinary(acc)
}

func Decode(accBytes []byte) (*Account, error) {
	ca := new(Account)
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

// Creates an otherwise zeroed Account from an Addressable and returns it as MutableAccount
func FromAddressable(addressable crypto.Addressable) *Account {
	return &Account{
		Address:   addressable.GetAddress(),
		PublicKey: addressable.GetPublicKey(),
		// Since nil slices and maps compare differently to empty ones
		Code: Bytecode{},
		Permissions: permission.AccountPermissions{
			Roles: []string{},
		},
	}
}

// Copies all mutable parts of account
func (acc *Account) Copy() *Account {
	if acc == nil {
		return nil
	}
	accCopy := *acc
	accCopy.Permissions.Roles = make([]string, len(acc.Permissions.Roles))
	copy(accCopy.Permissions.Roles, acc.Permissions.Roles)
	return &accCopy
}

func (acc *Account) Equal(accOther *Account) bool {
	accEnc, err := acc.Encode()
	if err != nil {
		return false
	}
	accOtherEnc, err := acc.Encode()
	if err != nil {
		return false
	}
	return bytes.Equal(accEnc, accOtherEnc)
}

func (acc Account) String() string {
	return fmt.Sprintf("Account{Address: %s; Sequence: %v; PublicKey: %v Balance: %v; CodeLength: %v; Permissions: %v}",
		acc.Address, acc.Sequence, acc.PublicKey, acc.Balance, len(acc.Code), acc.Permissions)
}

func (acc *Account) Tagged() query.Tagged {
	return &TaggedAccount{
		Account: acc,
		Tagged:  query.MustReflectTags(acc),
	}
}

type TaggedAccount struct {
	*Account
	query.Tagged
}
