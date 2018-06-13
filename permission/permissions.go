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
	"github.com/hyperledger/burrow/permission/types"
)

var (
	ZeroBasePermissions    = types.BasePermissions{0, 0}
	ZeroAccountPermissions = types.AccountPermissions{
		Base: ZeroBasePermissions,
	}
	DefaultAccountPermissions = types.AccountPermissions{
		Base: types.BasePermissions{
			Perms:  types.DefaultPermFlags,
			SetBit: types.AllPermFlags,
		},
		Roles: []string{},
	}
	AllAccountPermissions = types.AccountPermissions{
		Base: types.BasePermissions{
			Perms:  types.AllPermFlags,
			SetBit: types.AllPermFlags,
		},
		Roles: []string{},
	}
)
