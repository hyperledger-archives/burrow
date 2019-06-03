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

import "fmt"

// Gets the permission value.
// ErrValueNotSet is returned if the permission's set bits are not all on,
// and should be caught by caller so the global permission can be fetched
func (bp BasePermissions) Get(ty PermFlag) (bool, error) {
	if ty == 0 {
		return false, ErrInvalidPermission(ty)
	}
	if !bp.IsSet(ty) {
		return false, ErrValueNotSet(ty)
	}
	return bp.Perms&ty == ty, nil
}

// Set a permission bit. Will set the permission's set bit to true.
func (bp *BasePermissions) Set(ty PermFlag, value bool) error {
	if ty == 0 {
		return ErrInvalidPermission(ty)
	}
	bp.SetBit |= ty
	if value {
		bp.Perms |= ty
	} else {
		bp.Perms &= ^ty
	}
	return nil
}

// Set the permission's set bits to false
func (bp *BasePermissions) Unset(ty PermFlag) error {
	if ty == 0 {
		return ErrInvalidPermission(ty)
	}
	bp.SetBit &= ^ty
	return nil
}

// Check if the permission is set
func (bp BasePermissions) IsSet(ty PermFlag) bool {
	if ty == 0 {
		return false
	}
	return bp.SetBit&ty == ty
}

// Returns the Perms PermFlag masked with SetBit bit field to give the resultant
// permissions enabled by this BasePermissions
func (bp BasePermissions) ResultantPerms() PermFlag {
	return bp.Perms & bp.SetBit
}

// Returns a BasePermission that matches any permissions set on this BasePermission
// and falls through to any permissions set on the bpFallthrough
func (bp BasePermissions) Compose(bpFallthrough BasePermissions) BasePermissions {
	return BasePermissions{
		// Combine set perm flags from bp with set perm flags in fallthrough NOT set in bp
		Perms:  (bp.Perms & bp.SetBit) | (bpFallthrough.Perms & (^bp.SetBit & bpFallthrough.SetBit)),
		SetBit: bp.SetBit | bpFallthrough.SetBit,
	}
}

func (bp BasePermissions) String() string {
	return fmt.Sprintf("Base: %b; Set: %b", bp.Perms, bp.SetBit)
}
