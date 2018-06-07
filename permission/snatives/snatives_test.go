package snatives

import (
	"testing"

	permission "github.com/hyperledger/burrow/permission/types"
	"github.com/stretchr/testify/assert"
)

func TestPermArgs_String(t *testing.T) {
	role := "foo"
	value := true
	permFlag := permission.AddRole | permission.RemoveRole
	permArgs := PermArgs{
		PermFlag:   permission.SetBase,
		Permission: &permFlag,
		Role:       &role,
		Value:      &value,
	}
	assert.Equal(t, "PermArgs{PermFlag: setBase, Permission: addRole | removeRole, Role: foo, Value: true}",
		permArgs.String())
}
