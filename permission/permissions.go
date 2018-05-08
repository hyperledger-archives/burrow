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

	"github.com/hyperledger/burrow/permission/types"
)

//------------------------------------------------------------------------------------------------

// Base permission references are like unix (the index is already bit shifted)
const (
	// Chain permissions.
	// These permissions grant the ability for accounts to perform certain transition within the execution package
	// Root is a reserved permission currently unused that may be used in the future to grant super-user privileges
	// for instance to a governance contract
	Root types.PermFlag = 1 << iota // 1
	// Send permits an account to issue a SendTx to transfer value from one account to another. Note that value can
	// still be transferred with a CallTx by specifying an Amount in the InputTx. Funding an account is the basic
	// prerequisite for an account to act in the system so is often used as a surrogate for 'account creation' when
	// sending to a unknown account - in order for this to be permitted the input account needs the CreateAccount
	// permission in addition.
	Send // 2
	// Call permits and account to issue a CallTx, which can be used to call (run) the code of an existing
	// account/contract (these are synonymous in Burrow/EVM). A CallTx can be used to create an account if it points to
	// a nil address - in order for an account to be permitted to do this the input (calling) account needs the
	// CreateContract permission in addition.
	Call // 4
	// CreateContract permits the input account of a CallTx to create a new contract/account when CallTx.Address is nil
	// and permits an executing contract in the EVM to create a new contract programmatically.
	CreateContract // 8
	// CreateAccount permits an input account of a SendTx to add value to non-existing (unfunded) accounts
	CreateAccount // 16
	// Bond is a reserved permission for making changes to the validator set - currently unused
	Bond // 32
	// Name permits manipulation of the name registry by allowing an account to issue a NameTx
	Name // 64

	// Moderator permissions.
	// These permissions concern the alteration of the chain permissions listed above. Each permission relates to a
	// particular canonical permission mutation or query function. When an account is granted a moderation permission
	// it is permitted to call that function. See snative.go for a marked-up description of what each function does.
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

	// Chain permissions strings
	RootString           string = "root"
	SendString                  = "send"
	CallString                  = "call"
	CreateContractString        = "createContract"
	CreateAccountString         = "createAccount"
	BondString                  = "bond"
	NameString                  = "name"

	// Moderator permissions strings
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

// Returns the string name of a single bit non-composite PermFlag, or otherwise UnknownString
// See BasePermissionsToStringList to generate a string representation of a composite PermFlag
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
