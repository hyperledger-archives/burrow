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
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-wire"
)

func TestAddress(t *testing.T) {
	bs := []byte{
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
	}
	addr, err := AddressFromBytes(bs)
	assert.NoError(t, err)
	word256 := addr.Word256()
	leadingZeroes := []byte{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
	}
	assert.Equal(t, leadingZeroes, word256[:12])
	addrFromWord256 := AddressFromWord256(word256)
	assert.Equal(t, bs, addrFromWord256[:])
	assert.Equal(t, addr, addrFromWord256)
}

func TestAccountSerialise(t *testing.T) {
	type AccountContainingStruct struct {
		Account Account
		ChainID string
	}

	// This test is really testing this go wire declaration in account.go

	acc := NewConcreteAccountFromSecret("Super Semi Secret")

	// Go wire cannot serialise a top-level interface type it needs to be a field or sub-field of a struct
	// at some depth. i.e. you MUST wrap an interface if you want it to be decoded (they do not document this)
	var accStruct = AccountContainingStruct{
		Account: acc.Account(),
		ChainID: "TestChain",
	}

	// We will write into this
	accStructOut := AccountContainingStruct{}

	// We must pass in a value type to read from (accStruct), but provide a pointer type to write into (accStructout
	wire.ReadBinaryBytes(wire.BinaryBytes(accStruct), &accStructOut)

	assert.Equal(t, accStruct, accStructOut)
}

func TestDecodeConcrete(t *testing.T) {
	concreteAcc := NewConcreteAccountFromSecret("Super Semi Secret")
	acc := concreteAcc.Account()
	encodedAcc := acc.Encode()
	concreteAccOut, err := DecodeConcrete(encodedAcc)
	require.NoError(t, err)
	assert.Equal(t, concreteAcc, *concreteAccOut)
	concreteAccOut, err = DecodeConcrete([]byte("flungepliffery munknut tolopops"))
	assert.Error(t, err)
}

func TestDecode(t *testing.T) {
	concreteAcc := NewConcreteAccountFromSecret("Super Semi Secret")
	acc := concreteAcc.Account()
	accOut, err := Decode(acc.Encode())
	assert.NoError(t, err)
	assert.Equal(t, concreteAcc, *AsConcreteAccount(accOut))

	accOut, err = Decode([]byte("flungepliffery munknut tolopops"))
	assert.Error(t, err)
	assert.Nil(t, accOut)
}

func TestMarshalJSON(t *testing.T) {
	concreteAcc := NewConcreteAccountFromSecret("Super Semi Secret")
	concreteAcc.Code = []byte{60, 23, 45}
	acc := concreteAcc.Account()
	bs, err := json.Marshal(acc)
	assert.Equal(t, `{"Address":"745BD6BE33020146E04FA0F293A41E389887DE86","PublicKey":{"type":"ed25519","data":"8CEBC16C166A0614AD7C8E330318E774E1A039321F17274DF12ABA3B1BFC773C"},"Balance":0,"Code":"3C172D","Sequence":0,"StorageRoot":"","Permissions":{"Base":{"Perms":0,"SetBit":0},"Roles":[]}}`,
		string(bs))
	assert.NoError(t, err)
}
