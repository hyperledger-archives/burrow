package permission

import (
	"testing"

	"github.com/hyperledger/burrow/permission/types"
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
	permStrings, err := BasePermissionsToStringList(allSetBasePermission(Root | HasRole | SetBase | Call))
	require.NoError(t, err)
	assert.Equal(t, []string{"root", "call", "setBase", "hasRole"}, permStrings)

	permStrings, err = BasePermissionsToStringList(allSetBasePermission(AllPermFlags))
	require.NoError(t, err)
	assert.Equal(t, []string{"root", "send", "call", "createContract", "createAccount", "bond", "name", "hasBase",
		"setBase", "unsetBase", "setGlobal", "hasRole", "addRole", "removeRole"}, permStrings)

	permStrings, err = BasePermissionsToStringList(allSetBasePermission(AllPermFlags + 1))
	assert.Error(t, err)
}

func TestBasePermissionsString(t *testing.T) {
	permissionString := BasePermissionsString(allSetBasePermission(AllPermFlags &^ Root))
	assert.Equal(t, "send | call | createContract | createAccount | bond | name | hasBase | "+
		"setBase | unsetBase | setGlobal | hasRole | addRole | removeRole", permissionString)
}

func allSetBasePermission(perms types.PermFlag) types.BasePermissions {
	return types.BasePermissions{
		Perms:  perms,
		SetBit: perms,
	}
}
