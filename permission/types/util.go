package types

// ConvertMapStringIntToPermissions converts a map of string-integer pairs and a slice of
// strings for the roles to an AccountPermissions type.  The integer needs to be greater
// than zero to set the permission.  For all unmentioned permissions the ZeroBasePermissions
// is defaulted to.
// TODO: [ben] re-evaluate the use of int for setting the permission.
func ConvertPermissionsMapAndRolesToAccountPermissions(permissions map[string]int, roles []string) (*AccountPermissions, error) {
	var err error
	accountPermissions := &AccountPermissions{}
	accountPermissions.Base, err = convertPermissionsMapStringIntToBasePermissions(permissions)
	if err != nil {
		return nil, err
	}
	accountPermissions.Roles = roles
	return accountPermissions, nil
}

// convertPermissionsMapStringIntToBasePermissions converts a map of string-integer pairs to
// BasePermissions.
func convertPermissionsMapStringIntToBasePermissions(permissions map[string]int) (BasePermissions, error) {
	// initialise basePermissions as ZeroBasePermissions
	basePermissions := ZeroBasePermissions

	for permissionName, value := range permissions {
		permissionsFlag, err := PermStringToFlag(permissionName)
		if err != nil {
			return basePermissions, err
		}
		// sets the permissions bitflag and the setbit flag for the permission.
		basePermissions.Set(permissionsFlag, value > 0)
	}

	return basePermissions, nil
}
