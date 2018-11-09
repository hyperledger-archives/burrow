package permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllPermissions(t *testing.T) {
	assert.Equal(t, AllPermFlags, DefaultPermFlags|AddRole|RemoveRole|SetBase|UnsetBase|Root|SetGlobal|Proposal)
}
