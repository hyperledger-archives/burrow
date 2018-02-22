package permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermArgs_String(t *testing.T) {
	role := "foo"
	value := true
	permission := AddRole | RemoveRole
	permArgs := PermArgs{
		PermFlag:   SetBase,
		Permission: &permission,
		Role:       &role,
		Value:      &value,
	}
	assert.Equal(t, "PermArgs{PermFlag: setBase, Permission: addRole | removeRole, Role: foo, Value: true}",
		permArgs.String())
}
