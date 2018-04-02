// Copyright 2017 Monax Industries Limited
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
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/permission/types"
)

// ConvertMapStringIntToPermissions converts a map of string-bool pairs and a slice of
// strings for the roles to an AccountPermissions type. If the value in the
// permissions map is true for a particular permission string then the permission
// will be set in the AccountsPermissions. For all unmentioned permissions the
// ZeroBasePermissions is defaulted to.
func ConvertPermissionsMapAndRolesToAccountPermissions(permissions map[string]bool,
	roles []string) (*types.AccountPermissions, error) {
	var err error
	accountPermissions := &types.AccountPermissions{}
	accountPermissions.Base, err = convertPermissionsMapStringIntToBasePermissions(permissions)
	if err != nil {
		return nil, err
	}
	accountPermissions.Roles = roles
	return accountPermissions, nil
}

// convertPermissionsMapStringIntToBasePermissions converts a map of string-bool
// pairs to BasePermissions.
func convertPermissionsMapStringIntToBasePermissions(permissions map[string]bool) (types.BasePermissions, error) {
	// initialise basePermissions as ZeroBasePermissions
	basePermissions := ZeroBasePermissions

	for permissionName, value := range permissions {
		permissionsFlag, err := PermStringToFlag(permissionName)
		if err != nil {
			return basePermissions, err
		}
		// sets the permissions bitflag and the setbit flag for the permission.
		basePermissions.Set(permissionsFlag, value)
	}

	return basePermissions, nil
}

// Builds a composite BasePermission by creating a PermFlag from permissions strings and
// setting them all
func BasePermissionsFromStringList(permissions []string) (types.BasePermissions, error) {
	permFlag, err := PermFlagFromStringList(permissions)
	if err != nil {
		return ZeroBasePermissions, err
	}
	return types.BasePermissions{
		Perms:  permFlag,
		SetBit: permFlag,
	}, nil
}

// Builds a composite PermFlag by mapping each permission string in permissions to its
// flag and composing them with binary or
func PermFlagFromStringList(permissions []string) (types.PermFlag, error) {
	var permFlag types.PermFlag
	for _, perm := range permissions {
		flag, err := PermStringToFlag(perm)
		if err != nil {
			return permFlag, err
		}
		permFlag |= flag
	}
	return permFlag, nil
}

// Builds a list of set permissions from a BasePermission by creating a list of permissions strings
// from the resultant permissions of basePermissions
func BasePermissionsToStringList(basePermissions types.BasePermissions) ([]string, error) {
	return PermFlagToStringList(basePermissions.ResultantPerms())
}

// Creates a list of individual permission flag strings from a possibly composite PermFlag
// by projecting out each bit and adding its permission string if it is set
func PermFlagToStringList(permFlag types.PermFlag) ([]string, error) {
	permStrings := make([]string, 0, NumPermissions)
	if permFlag > AllPermFlags {
		return nil, fmt.Errorf("resultant permission 0b%b is invalid: has permission flag set above top flag 0b%b",
			permFlag, TopPermFlag)
	}
	for i := uint(0); i < NumPermissions; i++ {
		permFlag := permFlag & (1 << i)
		if permFlag > 0 {
			permStrings = append(permStrings, PermFlagToString(permFlag))
		}
	}
	return permStrings, nil
}

// Generates a human readable string from the resultant permissions of basePermission
func BasePermissionsString(basePermissions types.BasePermissions) string {
	permStrings, err := BasePermissionsToStringList(basePermissions)
	if err != nil {
		return UnknownString
	}
	return strings.Join(permStrings, " | ")
}

func String(permFlag types.PermFlag) string {
	permStrings, err := PermFlagToStringList(permFlag)
	if err != nil {
		return UnknownString
	}
	return strings.Join(permStrings, " | ")
}
