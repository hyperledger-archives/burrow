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
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission/types"
)

//---------------------------------------------------------------------------------------------------
// PermissionsTx.PermArgs interface and argument encoding

type PermArgs struct {
	PermFlag   types.PermFlag
	Address    acm.Address    `json:",omitempty"`
	Permission types.PermFlag `json:",omitempty"`
	Role       string         `json:",omitempty"`
	Value      bool           `json:",omitempty"`
}

func HasBaseArgs(address acm.Address, permission types.PermFlag) *PermArgs {
	return &PermArgs{
		PermFlag:   HasBase,
		Address:    address,
		Permission: permission,
	}
}

func SetBaseArgs(address acm.Address, permission types.PermFlag, value bool) *PermArgs {
	return &PermArgs{
		PermFlag:   SetBase,
		Address:    address,
		Permission: permission,
		Value:      value,
	}
}

func UnsetBaseArgs(address acm.Address, permission types.PermFlag) *PermArgs {
	return &PermArgs{
		PermFlag:   UnsetBase,
		Address:    address,
		Permission: permission,
	}
}

func SetGlobalArgs(permission types.PermFlag, value bool) *PermArgs {
	return &PermArgs{
		PermFlag:   SetGlobal,
		Permission: permission,
		Value:      value,
	}
}

func HasRoleArgs(address acm.Address, role string) *PermArgs {
	return &PermArgs{
		PermFlag: HasRole,
		Address:  address,
		Role:     role,
	}
}

func AddRoleArgs(address acm.Address, role string) *PermArgs {
	return &PermArgs{
		PermFlag: AddRole,
		Address:  address,
		Role:     role,
	}
}

func RemoveRoleArgs(address acm.Address, role string) *PermArgs {
	return &PermArgs{
		PermFlag: RemoveRole,
		Address:  address,
		Role:     role,
	}
}
