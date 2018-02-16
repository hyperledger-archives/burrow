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

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission/types"
)

//------------------------------------------------------------------------------------------------

// Base permission references are like unix (the index is already bit shifted)
const (
	// chain permissions
	Root           types.PermFlag = 1 << iota // 1
	Send                                      // 2
	Call                                      // 4
	CreateContract                            // 8
	CreateAccount                             // 16
	Bond                                      // 32
	Name                                      // 64

	// moderator permissions
	HasBase
	SetBase
	UnsetBase
	SetGlobal
	HasRole
	AddRole
	RemoveRole

	NumPermissions uint = 14 // NOTE Adjust this too. We can support upto 64

	TopPermFlag      types.PermFlag = 1 << (NumPermissions - 1)
	AllPermFlags     types.PermFlag = TopPermFlag | (TopPermFlag - 1)
	DefaultPermFlags types.PermFlag = Send | Call | CreateContract | CreateAccount | Bond | Name | HasBase | HasRole

	RootString           string = "root"
	SendString                  = "send"
	CallString                  = "call"
	CreateContractString        = "createContract"
	CreateAccountString         = "createAccount"
	BondString                  = "bond"
	NameString                  = "name"

	// moderator permissions
	HasBaseString    = "hasBase"
	SetBaseString    = "setBase"
	UnsetBaseString  = "unsetBase"
	SetGlobalString  = "setGlobal"
	HasRoleString    = "hasRole"
	AddRoleString    = "addRole"
	RemoveRoleString = "removeRole"
	UnknownString    = "#-UNKNOWN-#"

	AllString = "all"
)

var (
	ZeroBasePermissions    = types.BasePermissions{0, 0}
	ZeroAccountPermissions = types.AccountPermissions{
		Base: ZeroBasePermissions,
	}
	DefaultAccountPermissions = types.AccountPermissions{
		Base: types.BasePermissions{
			Perms:  DefaultPermFlags,
			SetBit: AllPermFlags,
		},
		Roles: []string{},
	}
	AllAccountPermissions = types.AccountPermissions{
		Base: types.BasePermissions{
			Perms:  AllPermFlags,
			SetBit: AllPermFlags,
		},
		Roles: []string{},
	}
)

//---------------------------------------------------------------------------------------------

//--------------------------------------------------------------------------------
// string utilities

// PermFlagToString assumes the permFlag is valid, else returns "#-UNKNOWN-#"
func PermFlagToString(pf types.PermFlag) string {
	switch pf {
	case AllPermFlags:
		return AllString
	case Root:
		return RootString
	case Send:
		return SendString
	case Call:
		return CallString
	case CreateContract:
		return CreateContractString
	case CreateAccount:
		return CreateAccountString
	case Bond:
		return BondString
	case Name:
		return NameString
	case HasBase:
		return HasBaseString
	case SetBase:
		return SetBaseString
	case UnsetBase:
		return UnsetBaseString
	case SetGlobal:
		return SetGlobalString
	case HasRole:
		return HasRoleString
	case AddRole:
		return AddRoleString
	case RemoveRole:
		return RemoveRoleString
	default:
		return UnknownString
	}
}

// PermStringToFlag maps camel- and snake case strings to the
// the corresponding permission flag.
func PermStringToFlag(perm string) (types.PermFlag, error) {
	switch strings.ToLower(perm) {
	case AllString:
		return AllPermFlags, nil
	case RootString:
		return Root, nil
	case SendString:
		return Send, nil
	case CallString:
		return Call, nil
	case CreateContractString, "createcontract", "create_contract":
		return CreateContract, nil
	case CreateAccountString, "createaccount", "create_account":
		return CreateAccount, nil
	case BondString:
		return Bond, nil
	case NameString:
		return Name, nil
	case HasBaseString, "hasbase", "has_base":
		return HasBase, nil
	case SetBaseString, "setbase", "set_base":
		return SetBase, nil
	case UnsetBaseString, "unsetbase", "unset_base":
		return UnsetBase, nil
	case SetGlobalString, "setglobal", "set_global":
		return SetGlobal, nil
	case HasRoleString, "hasrole", "has_role":
		return HasRole, nil
	case AddRoleString, "addrole", "add_role":
		return AddRole, nil
	case RemoveRoleString, "removerole", "rmrole", "rm_role":
		return RemoveRole, nil
	default:
		return 0, fmt.Errorf("unknown permission %s", perm)
	}
}

func GlobalPermissionsAccount(state acm.Getter) acm.Account {
	acc, err := state.GetAccount(GlobalPermissionsAddress)
	if err != nil {
		panic("Could not get global permission account, but this must exist")
	}
	return acc
}

// Get global permissions from the account at GlobalPermissionsAddress
func GlobalAccountPermissions(state acm.Getter) types.AccountPermissions {
	return GlobalPermissionsAccount(state).Permissions()
}
