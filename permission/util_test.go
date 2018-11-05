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
	assert.Equal(t, []string{"root", "send", "call", "createContract", "createAccount", "bond", "name", "proposal", "input", "hasBase",
		"setBase", "unsetBase", "setGlobal", "hasRole", "addRole", "removeRole"}, permStrings)

	permStrings = BasePermissionsToStringList(allSetBasePermission(AllPermFlags + 1))
	assert.Equal(t, []string{}, permStrings)
}

func TestBasePermissionsString(t *testing.T) {
	permissionString := BasePermissionsString(allSetBasePermission(AllPermFlags &^ Root))
	assert.Equal(t, "send | call | createContract | createAccount | bond | name | proposal | input | hasBase | "+
		"setBase | unsetBase | setGlobal | hasRole | addRole | removeRole", permissionString)
}

func allSetBasePermission(perms PermFlag) BasePermissions {
	return BasePermissions{
		Perms:  perms,
		SetBit: perms,
	}
}
