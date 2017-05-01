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

// TODO: [ben] Account and PrivateAccount need to become a pure interface
// and then move the implementation to the manager types.
// Eg, Geth has its accounts, different from BurrowMint

import (
	"bytes"
	"fmt"
	"io"

	"github.com/hyperledger/burrow/common/sanity"
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
		sanity.PanicCrisis(err)
	}

	return buf.Bytes()
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
