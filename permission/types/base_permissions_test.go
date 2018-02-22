package types

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasePermissions_Get(t *testing.T) {
	allPermsString := "11111111111111"

	hasPermission, err := testHasPermission(t,
		"100001000110",
		allPermsString,
		"100001000110")
	assert.NoError(t, err)
	assert.True(t, hasPermission)

	hasPermission, err = testHasPermission(t,
		allPermsString,
		allPermsString,
		"100000000")
	assert.NoError(t, err)
	assert.True(t, hasPermission)

	unset := "00100001000111"
	hasPermission, err = testHasPermission(t,
		unset,
		"11011110111000",
		unset)
	assert.Equal(t, ErrValueNotSet(PermFlagFromString(t, unset)), err)
	assert.False(t, hasPermission)

	hasPermission, err = testHasPermission(t,
		"00100001000111",
		"11011110111000",
		"00100001000111")
	assert.Error(t, err)
	assert.False(t, hasPermission)
}

func TestBasePermissions_Compose(t *testing.T) {
	assertComposition(t,
		"101010",
		"111000",
		"111111",
		"111111",
		"101111")

	assertComposition(t,
		"101010",
		"111111",
		"111111",
		"111111",
		"101010")

	assertComposition(t,
		"101010",
		"000001",
		"111111",
		"000000",
		"000000")

	assertComposition(t,
		"101010",
		"000001",
		"111111",
		"001100",
		"001100")

	assertComposition(t,
		"101010",
		"000101",
		"111111",
		"001100",
		"001000")

	assertComposition(t,
		"000000",
		"010101",
		"111111",
		"111111",
		"101010")
}

func assertComposition(t *testing.T, perms, setBit, permsFallback, setBitFallback, permsResultant string) {
	composed := BasePermissionsFromStrings(t, perms, setBit).
		Compose(BasePermissionsFromStrings(t, permsFallback, setBitFallback)).ResultantPerms()
	expected := PermFlagFromString(t, permsResultant)
	if !assert.Equal(t, expected, composed) {
		t.Errorf("\nexpected: %014b\nactual:   %014b", expected, composed)
	}
}

func testHasPermission(t *testing.T, perms, setBit, permsToCheck string) (bool, error) {
	return BasePermissionsFromStrings(t, perms, setBit).Get(PermFlagFromString(t, permsToCheck))
}

func BasePermissionsFromStrings(t *testing.T, perms, setBit string) BasePermissions {
	return BasePermissions{
		Perms:  PermFlagFromString(t, perms),
		SetBit: PermFlagFromString(t, setBit),
	}
}

func PermFlagFromString(t *testing.T, binaryString string) PermFlag {
	permFlag, err := strconv.ParseUint(binaryString, 2, 64)
	require.NoError(t, err)
	return PermFlag(permFlag)
}
