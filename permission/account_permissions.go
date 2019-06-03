// Copyright 2019 Monax Industries Limited
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

import "github.com/hyperledger/burrow/binary"

func NewAccountPermissions(pss ...PermFlag) AccountPermissions {
	var perms PermFlag
	for _, ps := range pss {
		perms |= ps
	}
	return AccountPermissions{
		Base: BasePermissions{
			Perms:  perms,
			SetBit: perms,
		},
	}
}

// Returns true if the role is found
func (ap AccountPermissions) HasRole(role string) bool {
	role = string(binary.RightPadBytes([]byte(role), 32))
	for _, r := range ap.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// Returns true if the role is added, and false if it already exists
func (ap *AccountPermissions) AddRole(role string) bool {
	role = string(binary.RightPadBytes([]byte(role), 32))
	for _, r := range ap.Roles {
		if r == role {
			return false
		}
	}
	ap.Roles = append(ap.Roles, role)
	return true
}

// Returns true if the role is removed, and false if it is not found
func (ap *AccountPermissions) RemoveRole(role string) bool {
	role = string(binary.RightPadBytes([]byte(role), 32))
	for i, r := range ap.Roles {
		if r == role {
			post := []string{}
			if len(ap.Roles) > i+1 {
				post = ap.Roles[i+1:]
			}
			ap.Roles = append(ap.Roles[:i], post...)
			return true
		}
	}
	return false
}

// Clone clones the account permissions
func (ap *AccountPermissions) Clone() AccountPermissions {
	// clone base permissions
	basePermissionsClone := ap.Base
	var rolesClone []string
	// It helps if we normalise empty roles to []string(nil) rather than []string{}
	if len(ap.Roles) > 0 {
		// clone roles []string
		rolesClone = make([]string, len(ap.Roles))
		// strings are immutable so copy suffices
		copy(rolesClone, ap.Roles)
	}

	return AccountPermissions{
		Base:  basePermissionsClone,
		Roles: rolesClone,
	}
}
