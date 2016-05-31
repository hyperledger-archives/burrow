// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

// TODO: [ben] Account and PrivateAccount need to become a pure interface
// and then move the implementation to the manager types.
// Eg, Geth has its accounts, different from ErisMint

package types

import (
	"bytes"
	"fmt"
	"io"

	ptypes "github.com/eris-ltd/eris-db/permission/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-merkle"
	"github.com/tendermint/go-wire"
)

// Signable is an interface for all signable things.
// It typically removes signatures before serializing.
type Signable interface {
	WriteSignBytes(chainID string, w io.Writer, n *int, err *error)
	// SignBytes is a convenience method for getting the bytes to sign of a Signable.
	SignBytes(chainID string, o Signable) []byte
}

// SignBytes is a convenience method for getting the bytes to sign of a Signable.
func SignBytes(chainID string, o Signable) []byte {
	buf, n, err := new(bytes.Buffer), new(int), new(error)
	o.WriteSignBytes(chainID, buf, n, err)
	if *err != nil {
		PanicCrisis(err)
	}
	return buf.Bytes()
}

// HashSignBytes is a convenience method for getting the hash of the bytes of a signable
func HashSignBytes(chainID string, o Signable) []byte {
	return merkle.SimpleHashFromBinary(SignBytes(chainID, o))
}

//-----------------------------------------------------------------------------

// Account resides in the application state, and is mutated by transactions
// on the blockchain.
// Serialized by wire.[read|write]Reflect
type Account struct {
	Address     []byte        `json:"address"`
	PubKey      crypto.PubKey `json:"pub_key"`
	Sequence    int           `json:"sequence"`
	Balance     int64         `json:"balance"`
	Code        []byte        `json:"code"`         // VM code
	StorageRoot []byte        `json:"storage_root"` // VM storage merkle root.

	Permissions ptypes.AccountPermissions `json:"permissions"`
}

func (acc *Account) Copy() *Account {
	accCopy := *acc
	return &accCopy
}

func (acc *Account) String() string {
	if acc == nil {
		return "nil-Account"
	}
	return fmt.Sprintf("Account{%X:%v B:%v C:%v S:%X P:%s}", acc.Address, acc.PubKey, acc.Balance, len(acc.Code), acc.StorageRoot, acc.Permissions)
}

func AccountEncoder(o interface{}, w io.Writer, n *int, err *error) {
	wire.WriteBinary(o.(*Account), w, n, err)
}

func AccountDecoder(r io.Reader, n *int, err *error) interface{} {
	return wire.ReadBinary(&Account{}, r, 0, n, err)
}

var AccountCodec = wire.Codec{
	Encode: AccountEncoder,
	Decode: AccountDecoder,
}

func EncodeAccount(acc *Account) []byte {
	w := new(bytes.Buffer)
	var n int
	var err error
	AccountEncoder(acc, w, &n, &err)
	return w.Bytes()
}

func DecodeAccount(accBytes []byte) *Account {
	var n int
	var err error
	acc := AccountDecoder(bytes.NewBuffer(accBytes), &n, &err)
	return acc.(*Account)
}
