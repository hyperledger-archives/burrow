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
	"math/rand"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
)

type deterministicGenesis struct {
	random *rand.Rand
}

// Generates deterministic pseudo-random genesis state
func NewDeterministicGenesis(seed int64) *deterministicGenesis {
	return &deterministicGenesis{
		random: rand.New(rand.NewSource(seed)),
	}
}

func (dg *deterministicGenesis) GenesisDoc(numAccounts int, randBalance bool, minBalance uint64, numValidators int,
	randBonded bool, minBonded int64) (*GenesisDoc, []*acm.PrivateAccount, []*acm.PrivateAccount) {

	accounts := make([]Account, numAccounts)
	privAccounts := make([]*acm.PrivateAccount, numAccounts)
	defaultPerms := permission.DefaultAccountPermissions
	for i := 0; i < numAccounts; i++ {
		account, privAccount := dg.Account(randBalance, minBalance)
		acc := Account{
			BasicAccount: BasicAccount{
				Address: account.GetAddress(),
				Amount:  account.Balance,
			},
			Permissions: defaultPerms.Clone(), // This will get copied into each state.Account.
		}
		acc.Permissions.Base.Set(permission.Root, true)
		accounts[i] = acc
		privAccounts[i] = privAccount
	}
	validators := make([]Validator, numValidators)
	privValidators := make([]*acm.PrivateAccount, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := acm.GeneratePrivateAccountFromSecret(fmt.Sprintf("val_%v", i))
		privValidators[i] = validator
		validators[i] = Validator{
			BasicAccount: BasicAccount{
				Address:   validator.GetAddress(),
				PublicKey: validator.GetPublicKey(),
				// Avoid max validator cap
				Amount: uint64(dg.random.Int63()/16 + 1),
			},
			UnbondTo: []BasicAccount{
				{
					Address: validator.GetAddress(),
					Amount:  uint64(dg.random.Int63()),
				},
			},
		}
	}
	return &GenesisDoc{
		ChainName:   "TestChain",
		GenesisTime: time.Unix(1506172037, 0).UTC(),
		Accounts:    accounts,
		Validators:  validators,
	}, privAccounts, privValidators

}

func (dg *deterministicGenesis) Account(randBalance bool, minBalance uint64) (*acm.Account, *acm.PrivateAccount) {
	privateKey, err := crypto.GeneratePrivateKey(dg.random, crypto.CurveTypeEd25519)
	if err != nil {
		panic(fmt.Errorf("could not generate private key deterministically"))
	}
	privAccount := &acm.ConcretePrivateAccount{
		PublicKey:  privateKey.GetPublicKey(),
		PrivateKey: privateKey,
		Address:    privateKey.GetPublicKey().GetAddress(),
	}
	perms := permission.DefaultAccountPermissions
	acc := &acm.Account{
		Address:     privAccount.Address,
		PublicKey:   privAccount.PublicKey,
		Sequence:    uint64(dg.random.Int()),
		Balance:     minBalance,
		Permissions: perms,
	}
	if randBalance {
		acc.Balance += uint64(dg.random.Int())
	}
	return acc, privAccount.PrivateAccount()
}
