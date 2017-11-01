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
	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission/types"
	"github.com/tendermint/go-wire"
)

//---------------------------------------------------------------------------------------------------
// PermissionsTx.PermArgs interface and argument encoding

// Arguments are a registered interface in the PermissionsTx,
// so binary handles the arguments and each permission function gets a type-byte
// PermFlag() maps the type-byte to the permission
// The account sending the PermissionsTx must have this PermFlag set
type PermArgs interface {
	PermFlag() types.PermFlag
}

const (
	PermArgsTypeHasBase   = byte(0x01)
	PermArgsTypeSetBase   = byte(0x02)
	PermArgsTypeUnsetBase = byte(0x03)
	PermArgsTypeSetGlobal = byte(0x04)
	PermArgsTypeHasRole   = byte(0x05)
	PermArgsTypeAddRole   = byte(0x06)
	PermArgsTypeRmRole    = byte(0x07)
)

// TODO: [ben] this registration needs to be lifted up
// and centralised in core; here it pulls in go-wire dependency
// while it suffices to have the type bytes defined;
// ---
// for wire.readReflect
var _ = wire.RegisterInterface(
	struct{ PermArgs }{},
	wire.ConcreteType{&HasBaseArgs{}, PermArgsTypeHasBase},
	wire.ConcreteType{&SetBaseArgs{}, PermArgsTypeSetBase},
	wire.ConcreteType{&UnsetBaseArgs{}, PermArgsTypeUnsetBase},
	wire.ConcreteType{&SetGlobalArgs{}, PermArgsTypeSetGlobal},
	wire.ConcreteType{&HasRoleArgs{}, PermArgsTypeHasRole},
	wire.ConcreteType{&AddRoleArgs{}, PermArgsTypeAddRole},
	wire.ConcreteType{&RmRoleArgs{}, PermArgsTypeRmRole},
)

type HasBaseArgs struct {
	Address    account.Address `json:"address"`
	Permission types.PermFlag  `json:"permission"`
}

func (*HasBaseArgs) PermFlag() types.PermFlag {
	return HasBase
}

type SetBaseArgs struct {
	Address    account.Address `json:"address"`
	Permission types.PermFlag  `json:"permission"`
	Value      bool            `json:"value"`
}

func (*SetBaseArgs) PermFlag() types.PermFlag {
	return SetBase
}

type UnsetBaseArgs struct {
	Address    account.Address `json:"address"`
	Permission types.PermFlag  `json:"permission"`
}

func (*UnsetBaseArgs) PermFlag() types.PermFlag {
	return UnsetBase
}

type SetGlobalArgs struct {
	Permission types.PermFlag `json:"permission"`
	Value      bool           `json:"value"`
}

func (*SetGlobalArgs) PermFlag() types.PermFlag {
	return SetGlobal
}

type HasRoleArgs struct {
	Address account.Address `json:"address"`
	Role    string          `json:"role"`
}

func (*HasRoleArgs) PermFlag() types.PermFlag {
	return HasRole
}

type AddRoleArgs struct {
	Address account.Address `json:"address"`
	Role    string          `json:"role"`
}

func (*AddRoleArgs) PermFlag() types.PermFlag {
	return AddRole
}

type RmRoleArgs struct {
	Address account.Address `json:"address"`
	Role    string          `json:"role"`
}

func (*RmRoleArgs) PermFlag() types.PermFlag {
	return RemoveRole
}
