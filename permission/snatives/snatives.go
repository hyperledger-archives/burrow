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

package snatives

import (
	"fmt"
	"strings"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/permission/types"
)

//---------------------------------------------------------------------------------------------------
// PermissionsTx.PermArgs interface and argument encoding

type PermArgs struct {
	PermFlag   types.PermFlag
	Address    *acm.Address    `json:",omitempty"`
	Permission *types.PermFlag `json:",omitempty"`
	Role       *string         `json:",omitempty"`
	Value      *bool           `json:",omitempty"`
}

func (pa PermArgs) String() string {
	body := make([]string, 0, 5)
	body = append(body, fmt.Sprintf("PermFlag: %s", permission.String(pa.PermFlag)))
	if pa.Address != nil {
		body = append(body, fmt.Sprintf("Address: %s", *pa.Address))
	}
	if pa.Permission != nil {
		body = append(body, fmt.Sprintf("Permission: %s", permission.String(*pa.Permission)))
	}
	if pa.Role != nil {
		body = append(body, fmt.Sprintf("Role: %s", *pa.Role))
	}
	if pa.Value != nil {
		body = append(body, fmt.Sprintf("Value: %v", *pa.Value))
	}
	return fmt.Sprintf("PermArgs{%s}", strings.Join(body, ", "))
}

func (pa PermArgs) EnsureValid() error {
	pf := pa.PermFlag
	// Address
	if pa.Address == nil && pf != permission.SetGlobal {
		return fmt.Errorf("PermArgs for PermFlag %v requires Address to be provided but was nil", pf)
	}
	if pf == permission.HasRole || pf == permission.AddRole || pf == permission.RemoveRole {
		// Role
		if pa.Role == nil {
			return fmt.Errorf("PermArgs for PermFlag %v requires Role to be provided but was nil", pf)
		}
		// Permission
	} else if pa.Permission == nil {
		return fmt.Errorf("PermArgs for PermFlag %v requires Permission to be provided but was nil", pf)
		// Value
	} else if (pf == permission.SetBase || pf == permission.SetGlobal) && pa.Value == nil {
		return fmt.Errorf("PermArgs for PermFlag %v requires Value to be provided but was nil", pf)
	}
	return nil
}

func HasBaseArgs(address acm.Address, permFlag types.PermFlag) PermArgs {
	return PermArgs{
		PermFlag:   permission.HasBase,
		Address:    &address,
		Permission: &permFlag,
	}
}

func SetBaseArgs(address acm.Address, permFlag types.PermFlag, value bool) PermArgs {
	return PermArgs{
		PermFlag:   permission.SetBase,
		Address:    &address,
		Permission: &permFlag,
		Value:      &value,
	}
}

func UnsetBaseArgs(address acm.Address, permFlag types.PermFlag) PermArgs {
	return PermArgs{
		PermFlag:   permission.UnsetBase,
		Address:    &address,
		Permission: &permFlag,
	}
}

func SetGlobalArgs(permFlag types.PermFlag, value bool) PermArgs {
	return PermArgs{
		PermFlag:   permission.SetGlobal,
		Permission: &permFlag,
		Value:      &value,
	}
}

func HasRoleArgs(address acm.Address, role string) PermArgs {
	return PermArgs{
		PermFlag: permission.HasRole,
		Address:  &address,
		Role:     &role,
	}
}

func AddRoleArgs(address acm.Address, role string) PermArgs {
	return PermArgs{
		PermFlag: permission.AddRole,
		Address:  &address,
		Role:     &role,
	}
}

func RemoveRoleArgs(address acm.Address, role string) PermArgs {
	return PermArgs{
		PermFlag: permission.RemoveRole,
		Address:  &address,
		Role:     &role,
	}
}
