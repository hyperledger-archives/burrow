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

package execution

import (
	"testing"

	acm "github.com/hyperledger/burrow/account"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tmlibs/db"
)

func TestState_UpdateAccount(t *testing.T) {
	state := NewState(db.NewMemDB())

	perm := ptypes.AccountPermissions{
		Base: ptypes.BasePermissions{
			Perms: ptypes.SetGlobal | ptypes.HasRole,
		},
		Roles: []string{},
	}
	account := acm.NewAccountFromSecret("Foo", perm)
	account.AddToBalance(100)

	err := state.UpdateAccount(account)
	err = state.Save()

	require.NoError(t, err)
	accountOut, err := state.GetAccount(account.Address())
	require.NoError(t, err)
	assert.Equal(t, account, accountOut)
}
