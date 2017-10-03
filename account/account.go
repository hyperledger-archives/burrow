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
	"bytes"
	"fmt"
	"io"

	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

// Signable is an interface for all signable things.
// It typically removes signatures before serializing.
type Signable interface {
	WriteSignBytes(chainID string, w io.Writer, n *int, err *error)
}

// SignBytes is a convenience method for getting the bytes to sign of a Signable.
func SignBytes(chainID string, o Signable) []byte {
	buf, n, err := new(bytes.Buffer), new(int), new(error)
	o.WriteSignBytes(chainID, buf, n, err)
	if *err != nil {
		panic(fmt.Sprintf("could not write sign bytes for a signable: %s", *err))
	}
	return buf.Bytes()
}

type Addressable interface {
	// Get the 20 byte EVM address of this account
	Address() Address
	// Public key from which the Address is derived
	PubKey() crypto.PubKey
}

// The default immutable interface to an account
type Account interface {
	Addressable
	// The value held by this account in terms of the chain-native token
	Balance() int64
	// The EVM byte code held by this account (or equivalently, this contract)
	Code() []byte
	// The sequence number (or nonce) of this account, incremented each time a mutation of the
	// Account is persisted to the blockchain state
	Sequence() int64
	// The hash of all the values in this accounts storage (typically the root of some subtree
	// in the merkle tree of global storage state)
	StorageRoot() []byte
	// The permission flags and roles for this account
	Permissions() ptypes.AccountPermissions
	// Obtain a deterministic serialisation of this account
	// (i.e. update order and Go runtime independent)
	Encode() []byte
}

type MutableAccount interface {
	Account
	// Set public key (needed for lazy initialisation), should also set the dependent address
	SetPubKey(pubKey crypto.PubKey) MutableAccount
	// Alter balance by amount which can be negative or positive
	AlterBalance(amount int64) MutableAccount
	// Set EVM byte code associated with account
	SetCode(code []byte) MutableAccount
	// Increment Sequence number by 1 (capturing the current Sequence number as the index for any pending mutations)
	IncSequence() MutableAccount
	// Set the storage root hash
	SetStorageRoot(storageRoot []byte) MutableAccount
	// Set account permissions
	SetPermissions(permissions ptypes.AccountPermissions) MutableAccount
	// Get a pointer this account's AccountPermissions in order to mutate them
	MutablePermissions() *ptypes.AccountPermissions
	// Create a complete copy of this MutableAccount that is itself mutable
	Copy() MutableAccount
}

// ConcreteAccount is the canonical serialisation and bash-in-place object for an Account
type ConcreteAccount struct {
	Address     Address                   `json:"address"`
	PubKey      crypto.PubKey             `json:"pub_key"`
	Balance     int64                     `json:"balance"`
	Code        []byte                    `json:"code"` // VM code
	Sequence    int64                     `json:"sequence"`
	StorageRoot []byte                    `json:"storage_root"` // VM storage merkle root.
	Permissions ptypes.AccountPermissions `json:"permissions"`
}

// Returns a mutable, serialisable ConcreteAccount by copying from account
func AsConcreteAccount(account Account) ConcreteAccount {
	if account == nil {
		return ConcreteAccount{}
	}
	return ConcreteAccount{
		Address:     account.Address(),
		PubKey:      account.PubKey(),
		Balance:     account.Balance(),
		Code:        account.Code(),
		Sequence:    account.Sequence(),
		StorageRoot: account.StorageRoot(),
		Permissions: account.Permissions(),
	}
}

// Creates an otherwise zeroed Account from an Addressable
func FromAddressable(addressable Addressable) Account {
	return ConcreteAccount{
		Address: addressable.Address(),
		PubKey:  addressable.PubKey(),
		// Since nil slices and maps compare differently to empty ones
		Code:        []byte{},
		StorageRoot: []byte{},
		Permissions: ptypes.AccountPermissions{
			Roles: []string{},
		},
	}.Account()
}

// Returns an immutable account by copying from account
func AsAccount(account Account) Account {
	if account == nil {
		return nil
	}
	return AsConcreteAccount(account).Account()
}

// Returns a MutableAccount by copying from account
func AsMutableAccount(account Account) MutableAccount {
	if account == nil {
		return nil
	}
	return AsConcreteAccount(account).MutableAccount()
}

// concreteAccountWrapper wraps ConcreteAccount to provide a immutable read-only view
// via its implementation of Account and a mutable implementation via its implementation of
// MutableAccount
type concreteAccountWrapper struct {
	*ConcreteAccount `json:"unwrap"`
}

var _ = wire.RegisterInterface(struct{ Account }{}, wire.ConcreteType{concreteAccountWrapper{}, 0x01})

// Immutable Account implementation
var _ Account = concreteAccountWrapper{}

func (caw concreteAccountWrapper) Address() Address {
	return caw.ConcreteAccount.Address
}

func (caw concreteAccountWrapper) PubKey() crypto.PubKey {
	return caw.ConcreteAccount.PubKey
}

func (caw concreteAccountWrapper) Balance() int64 {
	return caw.ConcreteAccount.Balance
}

func (caw concreteAccountWrapper) Code() []byte {
	return caw.ConcreteAccount.Code
}

func (caw concreteAccountWrapper) Sequence() int64 {
	return caw.ConcreteAccount.Sequence
}

func (caw concreteAccountWrapper) StorageRoot() []byte {
	return caw.ConcreteAccount.StorageRoot
}

func (caw concreteAccountWrapper) Permissions() ptypes.AccountPermissions {
	return caw.ConcreteAccount.Permissions
}

func (caw concreteAccountWrapper) Encode() []byte {
	return caw.ConcreteAccount.Encode()
}

// Account mutation via MutableAccount interface
var _ MutableAccount = concreteAccountWrapper{}

func (caw concreteAccountWrapper) SetPubKey(pubKey crypto.PubKey) MutableAccount {
	caw.ConcreteAccount.PubKey = pubKey
	addressFromPubKey, err := AddressFromBytes(pubKey.Address())
	if err != nil {
		// We rely on this working in all over the place so shouldn't happen
		panic(fmt.Errorf("could not obtain address from public key: %v", pubKey))
	}
	// We don't want the wrong public key to take control of an account so we panic here
	if caw.ConcreteAccount.Address != addressFromPubKey {
		panic(fmt.Errorf("attempt to set public key of account %s to %v, "+
			"but that public key has address %s",
			caw.ConcreteAccount.Address, pubKey, addressFromPubKey))
	}
	return caw
}

func (caw concreteAccountWrapper) AlterBalance(amount int64) MutableAccount {
	caw.ConcreteAccount.Balance += amount
	return caw
}

func (caw concreteAccountWrapper) SetCode(code []byte) MutableAccount {
	caw.ConcreteAccount.Code = code
	return caw
}

func (caw concreteAccountWrapper) IncSequence() MutableAccount {
	caw.ConcreteAccount.Sequence += 1
	return caw
}

func (caw concreteAccountWrapper) SetStorageRoot(storageRoot []byte) MutableAccount {
	caw.ConcreteAccount.StorageRoot = storageRoot
	return caw
}

func (caw concreteAccountWrapper) SetPermissions(permissions ptypes.AccountPermissions) MutableAccount {
	caw.ConcreteAccount.Permissions = permissions
	return caw
}

func (caw concreteAccountWrapper) MutablePermissions() *ptypes.AccountPermissions {
	return &caw.ConcreteAccount.Permissions
}

func (caw concreteAccountWrapper) Copy() MutableAccount {
	return concreteAccountWrapper{caw.ConcreteAccount.Copy()}
}

// Return as immutable Account
func (acc ConcreteAccount) Account() Account {
	return concreteAccountWrapper{&acc}
}

// Return as mutable MutableACCount
func (acc ConcreteAccount) MutableAccount() MutableAccount {
	return concreteAccountWrapper{&acc}
}

func (acc *ConcreteAccount) Encode() []byte {
	w := new(bytes.Buffer)
	var n int
	var err error
	AccountEncoder(acc, w, &n, &err)
	return w.Bytes()
}

func (acc *ConcreteAccount) Copy() *ConcreteAccount {
	accCopy := *acc
	return &accCopy
}

func (acc *ConcreteAccount) String() string {
	if acc == nil {
		return "nil-Account"
	}
	return fmt.Sprintf("Account{%s:%v B:%v C:%v S:%X P:%s}", acc.Address, acc.PubKey, acc.Balance,
		len(acc.Code), acc.StorageRoot, acc.Permissions)
}

func AccountEncoder(o interface{}, w io.Writer, n *int, err *error) {
	wire.WriteBinary(o.(*ConcreteAccount), w, n, err)
}

func AccountDecoder(r io.Reader, n *int, err *error) interface{} {
	return wire.ReadBinary(&ConcreteAccount{}, r, 0, n, err)
}

var AccountCodec = wire.Codec{
	Encode: AccountEncoder,
	Decode: AccountDecoder,
}

func Decode(accBytes []byte) Account {
	ca, err := DecodeConcrete(accBytes)
	if err != nil {
		return nil
	}
	return ca.Account()
}

func DecodeConcrete(accBytes []byte) (*ConcreteAccount, error) {
	var n int
	var err error
	account := AccountDecoder(bytes.NewBuffer(accBytes), &n, &err)
	if err != nil {
		return nil, err
	}
	ca, ok := account.(*ConcreteAccount)
	if !ok {
		return nil, fmt.Errorf("could not convert account to *ConcreteAccount")
	}
	return ca, nil
}
