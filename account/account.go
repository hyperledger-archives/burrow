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
	Address() Address
	PubKey() crypto.PubKey
}

type Account interface {
	Addressable
	Balance() int64
	Code() []byte
	Sequence() int64
	StorageRoot() []byte
	Permissions() ptypes.AccountPermissions
}

// ConcreteAccount is the canonical serialisation object for Account
type ConcreteAccount struct {
	Address     Address                   `json:"address"`
	PubKey      crypto.PubKey             `json:"pub_key"`
	Balance     int64                     `json:"balance"`
	Code        []byte                    `json:"code"` // VM code
	Sequence    int64                     `json:"sequence"`
	StorageRoot []byte                    `json:"storage_root"` // VM storage merkle root.
	Permissions ptypes.AccountPermissions `json:"permissions"`
}

type concreteAccountWrapper struct {
	*ConcreteAccount
}

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

func (caw concreteAccountWrapper) Unwrap() *ConcreteAccount {
	return caw.ConcreteAccount
}

func (acc *ConcreteAccount) Wrap() concreteAccountWrapper {
	return concreteAccountWrapper{acc}
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

func EncodeAccount(acc *ConcreteAccount) []byte {
	w := new(bytes.Buffer)
	var n int
	var err error
	AccountEncoder(acc, w, &n, &err)
	return w.Bytes()
}

func DecodeAccount(accBytes []byte) *ConcreteAccount {
	var n int
	var err error
	acc := AccountDecoder(bytes.NewBuffer(accBytes), &n, &err)
	return acc.(*ConcreteAccount)
}
