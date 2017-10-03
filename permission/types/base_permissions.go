package types

import "fmt"

// A particular permission
type PermFlag uint64

// permission number out of bounds
type ErrInvalidPermission PermFlag

func (e ErrInvalidPermission) Error() string {
	return fmt.Sprintf("invalid permission %d", e)
}

// set=false. This error should be caught and the global
// value fetched for the permission by the caller
type ErrValueNotSet PermFlag

func (e ErrValueNotSet) Error() string {
	return fmt.Sprintf("the value for permission %d is not set", e)
}

// Base chain permissions struct
type BasePermissions struct {
	// bit array with "has"/"doesn't have" for each permission
	Perms PermFlag `json:"perms"`

	// bit array with "set"/"not set" for each permission (not-set should fall back to global)
	SetBit PermFlag `json:"set"`
}

// Get a permission value. ty should be a power of 2.
// ErrValueNotSet is returned if the permission's set bit is off,
// and should be caught by caller so the global permission can be fetched
func (p BasePermissions) Get(ty PermFlag) (bool, error) {
	if ty == 0 {
		return false, ErrInvalidPermission(ty)
	}
	if p.SetBit&ty == 0 {
		return false, ErrValueNotSet(ty)
	}
	return p.Perms&ty > 0, nil
}

// Set a permission bit. Will set the permission's set bit to true.
func (p *BasePermissions) Set(ty PermFlag, value bool) error {
	if ty == 0 {
		return ErrInvalidPermission(ty)
	}
	p.SetBit |= ty
	if value {
		p.Perms |= ty
	} else {
		p.Perms &= ^ty
	}
	return nil
}

// Set the permission's set bit to false
func (p *BasePermissions) Unset(ty PermFlag) error {
	if ty == 0 {
		return ErrInvalidPermission(ty)
	}
	p.SetBit &= ^ty
	return nil
}

// Check if the permission is set
func (p BasePermissions) IsSet(ty PermFlag) bool {
	if ty == 0 {
		return false
	}
	return p.SetBit&ty > 0
}

// Returns the Perms PermFlag masked with SetBit bit field to give the resultant
// permissions enabled by this BasePermissions
func (p BasePermissions) ResultantPerms() PermFlag {
	return p.Perms & p.SetBit
}

func (p BasePermissions) String() string {
	return fmt.Sprintf("Base: %b; Set: %b", p.Perms, p.SetBit)
}

//---------------------------------------------------------------------------------------------
