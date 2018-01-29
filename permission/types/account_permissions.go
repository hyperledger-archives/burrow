package types

import "github.com/hyperledger/burrow/binary"

type AccountPermissions struct {
	Base  BasePermissions
	Roles []string
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
func (ap *AccountPermissions) RmRole(role string) bool {
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
	// clone roles []string
	rolesClone := make([]string, len(ap.Roles))
	// strings are immutable so copy suffices
	copy(rolesClone, ap.Roles)

	return AccountPermissions{
		Base:  basePermissionsClone,
		Roles: rolesClone,
	}
}
