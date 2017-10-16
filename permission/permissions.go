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
	RmRole

	NumPermissions uint = 14 // NOTE Adjust this too. We can support upto 64

	TopPermFlag      types.PermFlag = 1 << (NumPermissions - 1)
	AllPermFlags     types.PermFlag = TopPermFlag | (TopPermFlag - 1)
	DefaultPermFlags types.PermFlag = Send | Call | CreateContract | CreateAccount | Bond | Name | HasBase | HasRole
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
func PermFlagToString(pf types.PermFlag) (perm string) {
	switch pf {
	case Root:
		perm = "root"
	case Send:
		perm = "send"
	case Call:
		perm = "call"
	case CreateContract:
		perm = "create_contract"
	case CreateAccount:
		perm = "create_account"
	case Bond:
		perm = "bond"
	case Name:
		perm = "name"
	case HasBase:
		perm = "hasBase"
	case SetBase:
		perm = "setBase"
	case UnsetBase:
		perm = "unsetBase"
	case SetGlobal:
		perm = "setGlobal"
	case HasRole:
		perm = "hasRole"
	case AddRole:
		perm = "addRole"
	case RmRole:
		perm = "removeRole"
	default:
		perm = "#-UNKNOWN-#"
	}
	return
}

// PermStringToFlag maps camel- and snake case strings to the
// the corresponding permission flag.
func PermStringToFlag(perm string) (pf types.PermFlag, err error) {
	switch strings.ToLower(perm) {
	case "root":
		pf = Root
	case "send":
		pf = Send
	case "call":
		pf = Call
	case "createcontract", "create_contract":
		pf = CreateContract
	case "createaccount", "create_account":
		pf = CreateAccount
	case "bond":
		pf = Bond
	case "name":
		pf = Name
	case "hasbase", "has_base":
		pf = HasBase
	case "setbase", "set_base":
		pf = SetBase
	case "unsetbase", "unset_base":
		pf = UnsetBase
	case "setglobal", "set_global":
		pf = SetGlobal
	case "hasrole", "has_role":
		pf = HasRole
	case "addrole", "add_role":
		pf = AddRole
	case "removerole", "rmrole", "rm_role":
		pf = RmRole
	default:
		err = fmt.Errorf("Unknown permission %s", perm)
	}
	return
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
