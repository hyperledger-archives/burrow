// Copyright 2015-2017 Monax Industries Limited.
// This file is part of the Monax platform (Monax)

// Monax is free software: you can use, redistribute it and/or modify
// it only under the terms of the GNU General Public License, version
// 3, as published by the Free Software Foundation.

// Monax is distributed WITHOUT ANY WARRANTY pursuant to
// the terms of the Gnu General Public Licence, version 3, including
// (but not limited to) Clause 15 thereof. See the text of the
// GNU General Public License, version 3 for full terms.

// You should have received a copy of the GNU General Public License,
// version 3, with Monax.  If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"github.com/tendermint/go-wire"
)

//---------------------------------------------------------------------------------------------------
// PermissionsTx.PermArgs interface and argument encoding

// Arguments are a registered interface in the PermissionsTx,
// so binary handles the arguments and each permission function gets a type-byte
// PermFlag() maps the type-byte to the permission
// The account sending the PermissionsTx must have this PermFlag set
type PermArgs interface {
	PermFlag() PermFlag
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
	Address    []byte   `json:"address"`
	Permission PermFlag `json:"permission"`
}

func (*HasBaseArgs) PermFlag() PermFlag {
	return HasBase
}

type SetBaseArgs struct {
	Address    []byte   `json:"address"`
	Permission PermFlag `json:"permission"`
	Value      bool     `json:"value"`
}

func (*SetBaseArgs) PermFlag() PermFlag {
	return SetBase
}

type UnsetBaseArgs struct {
	Address    []byte   `json:"address"`
	Permission PermFlag `json:"permission"`
}

func (*UnsetBaseArgs) PermFlag() PermFlag {
	return UnsetBase
}

type SetGlobalArgs struct {
	Permission PermFlag `json:"permission"`
	Value      bool     `json:"value"`
}

func (*SetGlobalArgs) PermFlag() PermFlag {
	return SetGlobal
}

type HasRoleArgs struct {
	Address []byte `json:"address"`
	Role    string `json:"role"`
}

func (*HasRoleArgs) PermFlag() PermFlag {
	return HasRole
}

type AddRoleArgs struct {
	Address []byte `json:"address"`
	Role    string `json:"role"`
}

func (*AddRoleArgs) PermFlag() PermFlag {
	return AddRole
}

type RmRoleArgs struct {
	Address []byte `json:"address"`
	Role    string `json:"role"`
}

func (*RmRoleArgs) PermFlag() PermFlag {
	return RmRole
}
