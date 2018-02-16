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
}
