// Copyright 2017 Monax Industries Limited
//.
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

package state

import (
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestState_UpdateAccount(t *testing.T) {
	s := NewState(dbm.NewMemDB())
	account := acm.NewAccountFromSecret("Foo")
	account.EVMCode = acm.Bytecode{1, 2, 3}
	account.Permissions.Base.Perms = permission.SetGlobal | permission.HasRole
	_, _, err := s.Update(func(ws Updatable) error {
		return ws.UpdateAccount(account)
	})
	require.NoError(t, err)

	require.NoError(t, err)
	accountOut, err := s.GetAccount(account.Address)
	require.NoError(t, err)
	assert.Equal(t, source.JSONString(account), source.JSONString(accountOut))
}
