// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package acm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/solidity"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
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

func TestMarshalJSON(t *testing.T) {
	acc := NewAccountFromSecret("Super Semi Secret")
	acc.EVMCode = NewEVMCode([]byte{60, 23, 45})
	acc.Permissions = permission.AccountPermissions{
		Base: permission.BasePermissions{
			Perms: permission.AllPermFlags,
		},
	}
	acc.Sequence = 4
	acc.Balance = 10
	bs, err := json.Marshal(acc)

	expected := fmt.Sprintf(`{"Address":"%s","PublicKey":{"CurveType":"ed25519","PublicKey":"%s"},"Sequence":4,"Balance":10,"EVMCode":{"Bytecode":"3C172D","OpcodeBitset":"Bw=="},"Permissions":{"Base":{"Perms":"root | send | call | createContract | createAccount | bond | name | proposal | input | batch | identify | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole","SetBit":""}}}`,
		acc.Address, acc.PublicKey)
	assert.Equal(t, expected, string(bs))
	assert.NoError(t, err)
}

func TestAccountTags(t *testing.T) {
	perms := permission.DefaultAccountPermissions
	perms.Roles = []string{"frogs", "dogs"}
	acc := &Account{
		Permissions: perms,
		EVMCode:     NewEVMCode(solidity.Bytecode_StrangeLoop),
	}
	flag, _ := acc.Get("Permissions")
	permString := permission.String(flag.(permission.PermFlag))
	assert.Equal(t, "send | call | createContract | createAccount | bond | name | proposal | input | batch | hasBase | hasRole", permString)
	roles, _ := acc.Get("Roles")
	assert.Equal(t, []string{"frogs", "dogs"}, roles)
	acc.Get("EVMCode")
	qry, err := query.New("EVMCode.Bytecode CONTAINS '0116002556001600360006101000A815'")
	require.NoError(t, err)
	assert.True(t, qry.Matches(acc))
}

func TestAccount_IsOpcodeAt(t *testing.T) {
	acc := Account{}
	// Fail safe before initialised by initEVMCode
	assert.False(t, acc.EVMCode.IsOpcode(0))
	assert.False(t, acc.EVMCode.IsOpcode(100))
}
