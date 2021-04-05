package permission

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllPermissions(t *testing.T) {
	assert.Equal(t, AllPermFlags, DefaultPermFlags|AddRole|RemoveRole|SetBase|UnsetBase|Root|SetGlobal|Proposal|Identify)
}

func TestName(t *testing.T) {
	fmt.Println(PermFlagToStringList(PermFlag(59007)))
	fmt.Println(PermFlagToStringList(PermFlag(8262)))
	fmt.Println(PermFlagToStringList(PermFlag(8263)))
	fmt.Println(PermFlagToStringList(DefaultPermFlags))
	fmt.Printf("%d\n", DefaultPermFlags)
}
