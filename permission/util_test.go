// Copyright 2019 Monax Industries Limited
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

package permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasePermissionsFromStringList(t *testing.T) {
	basePerms, err := BasePermissionsFromStringList([]string{HasRoleString, CreateContractString, SendString})
	require.NoError(t, err)
	permFlag := HasRole | CreateContract | Send
	assert.Equal(t, permFlag, basePerms.Perms)
	assert.Equal(t, permFlag, basePerms.SetBit)

	basePerms, err = BasePermissionsFromStringList([]string{AllString})
	require.NoError(t, err)
	permFlag = AllPermFlags
	assert.Equal(t, permFlag, basePerms.Perms)
	assert.Equal(t, permFlag, basePerms.SetBit)

	basePerms, err = BasePermissionsFromStringList([]string{"justHaveALittleRest"})
	assert.Error(t, err)
}

func TestBasePermissionsToStringList(t *testing.T) {
	permStrings := BasePermissionsToStringList(allSetBasePermission(Root | HasRole | SetBase | Call))
	assert.Equal(t, []string{"root", "call", "setBase", "hasRole"}, permStrings)

	permStrings = BasePermissionsToStringList(allSetBasePermission(AllPermFlags))
	assert.Equal(t, []string{"root", "send", "call", "createContract", "createAccount", "bond", "name", "proposal", "input", "batch", "hasBase",
		"setBase", "unsetBase", "setGlobal", "hasRole", "addRole", "removeRole"}, permStrings)

	permStrings = BasePermissionsToStringList(allSetBasePermission(AllPermFlags + 1))
	assert.Equal(t, []string{}, permStrings)
}

func TestBasePermissionsString(t *testing.T) {
	permissionString := BasePermissionsString(allSetBasePermission(AllPermFlags &^ Root))
	assert.Equal(t, "send | call | createContract | createAccount | bond | name | proposal | input | batch | hasBase | "+
		"setBase | unsetBase | setGlobal | hasRole | addRole | removeRole", permissionString)
}

func allSetBasePermission(perms PermFlag) BasePermissions {
	return BasePermissions{
		Perms:  perms,
		SetBit: perms,
	}
}
