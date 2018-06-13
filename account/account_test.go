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
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/permission/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddress(t *testing.T) {
	bs := []byte{
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
		1, 2, 3, 4, 5,
	}
	addr, err := crypto.AddressFromBytes(bs)
	assert.NoError(t, err)
	word256 := addr.Word256()
	leadingZeroes := []byte{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
	}
	assert.Equal(t, leadingZeroes, word256[:12])
	addrFromWord256 := crypto.AddressFromWord256(word256)
	assert.Equal(t, bs, addrFromWord256[:])
	assert.Equal(t, addr, addrFromWord256)
}

func TestDecodeAccount(t *testing.T) {
	permissions := types.AccountPermissions{
		Base: types.BasePermissions{
			Perms:  permission.SetGlobal,
			SetBit: permission.SetGlobal,
		},
		Roles: []string{"bums"},
	}

	acc := NewAccountFromSecret("Super Semi Secret", permissions)
	acc.AddToBalance(100)

	encodedAcc, err := acc.Encode()
	require.NoError(t, err)
	accOut, err := Decode(encodedAcc)
	require.NoError(t, err)
	assert.Equal(t, acc, accOut)

	accOut, err = Decode([]byte("flungepliffery munknut tolopops"))
	require.Error(t, err)
	assert.Nil(t, accOut)
}

func TestMarshalJSON(t *testing.T) {
	acc := NewContractAccountFromSecret("Super Semi Secret", permission.ZeroAccountPermissions)
	acc.AddToBalance(100)
	acc.SetCode([]byte{60, 23, 45})

	bs, err := json.Marshal(acc)

	expected := fmt.Sprintf(`{"Address":"%s","PublicKey":{"CurveType":"secp256k1","PublicKey":"%s"},`+
		`"Sequence":0,"Balance":100,"Code":"3C172D","StorageRoot":"",`+
		`"Permissions":{"Base":{"Perms":0,"SetBit":0},"Roles":[]}}`,
		acc.Address(), acc.PublicKey())
	assert.Equal(t, expected, string(bs))
	assert.NoError(t, err)
}
