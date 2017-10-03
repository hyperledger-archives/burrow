package types

import "github.com/hyperledger/burrow/word"

type AccountPermissions struct {
	Base  BasePermissions `json:"base"`
	Roles []string        `json:"roles"`
}

// Returns true if the role is found
func (aP AccountPermissions) HasRole(role string) bool {
	role = string(word.RightPadBytes([]byte(role), 32))
	for _, r := range aP.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// Returns true if the role is added, and false if it already exists
func (aP *AccountPermissions) AddRole(role string) bool {
	role = string(word.RightPadBytes([]byte(role), 32))
	for _, r := range aP.Roles {
		if r == role {
			return false
		}
	}
	aP.Roles = append(aP.Roles, role)
	return true
}

// Returns true if the role is removed, and false if it is not found
func (aP *AccountPermissions) RmRole(role string) bool {
	role = string(word.RightPadBytes([]byte(role), 32))
	for i, r := range aP.Roles {
		if r == role {
			post := []string{}
			if len(aP.Roles) > i+1 {
				post = aP.Roles[i+1:]
			}
			aP.Roles = append(aP.Roles[:i], post...)
			return true
		}
	}
	return false
}

// Clone clones the account permissions
func (accountPermissions *AccountPermissions) Clone() AccountPermissions {
	// clone base permissions
	basePermissionsClone := accountPermissions.Base
	// clone roles []string
	rolesClone := make([]string, len(accountPermissions.Roles))
	// strings are immutable so copy suffices
	copy(rolesClone, accountPermissions.Roles)

	return AccountPermissions{
		Base:  basePermissionsClone,
		Roles: rolesClone,
	}
}
