package permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermArgs_String(t *testing.T) {
	role := "foo"
	value := true
	permission := AddRole
	permArgs := PermArgs{
		PermFlag:   SetBase,
		Permission: &permission,
		Role:       &role,
		Value:      &value,
	}
	assert.Equal(t, "PermArgs{PermFlag: 0b100000000, Permission: 0b1000000000000, Role: foo, Value: true}",
		permArgs.String())
}
