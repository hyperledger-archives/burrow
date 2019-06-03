// Copyright 2019 Monax Industries Limited
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

package genesis

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
)

var genesisTime, _ = time.Parse("02-01-2006", "27-10-2017")

func TestMakeGenesisDocFromAccounts(t *testing.T) {
	genDoc := MakeGenesisDocFromAccounts("test-chain", nil, genesisTime,
		accountMap("Tinkie-winkie", "Lala", "Po", "Dipsy"),
		validatorMap("Foo", "Bar", "Baz"),
	)

	// Check we have matching serialisation after a round trip
	bs, err := genDoc.JSONBytes()
	assert.NoError(t, err)

	genDocOut, err := GenesisDocFromJSON(bs)
	assert.NoError(t, err)

	bsOut, err := genDocOut.JSONBytes()
	assert.NoError(t, err)

	assert.Equal(t, bs, bsOut)
	assert.Equal(t, genDoc.Hash(), genDocOut.Hash())
	fmt.Println(string(bs))
}

func accountMap(names ...string) map[string]*acm.Account {
	accounts := make(map[string]*acm.Account, len(names))
	for _, name := range names {
		accounts[name] = accountFromName(name)
	}
	return accounts
}

func validatorMap(names ...string) map[string]*validator.Validator {
	validators := make(map[string]*validator.Validator, len(names))
	for _, name := range names {
		acc := accountFromName(name)
		validators[name] = validator.FromAccount(acc, acc.Balance)
	}
	return validators
}

func accountFromName(name string) *acm.Account {
	ca := acm.NewAccountFromSecret(name)
	for _, c := range name {
		ca.Balance += uint64(c)
	}
	ca.Permissions = permission.AllAccountPermissions.Clone()
	return ca
}
