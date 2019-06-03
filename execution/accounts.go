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

package execution

import (
	"sync"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys"
	burrow_sync "github.com/hyperledger/burrow/sync"
)

// Accounts pairs an underlying state.Reader with a KeyClient to provide a signing variant of an account
// it also maintains a lock over addresses to provide a linearisation of signing events using SequentialSigningAccount
type Accounts struct {
	burrow_sync.RingMutex
	acmstate.Reader
	keyClient keys.KeyClient
}

type SigningAccount struct {
	*acm.Account
	crypto.Signer
}

type SequentialSigningAccount struct {
	Address       crypto.Address
	accountLocker sync.Locker
	getter        func() (*SigningAccount, error)
}

func NewAccounts(reader acmstate.Reader, keyClient keys.KeyClient, mutexCount int) *Accounts {
	return &Accounts{
		RingMutex: *burrow_sync.NewRingMutexNoHash(mutexCount),
		Reader:    reader,
		keyClient: keyClient,
	}
}
func (accs *Accounts) SigningAccount(address crypto.Address) (*SigningAccount, error) {
	signer, err := keys.AddressableSigner(accs.keyClient, address)
	if err != nil {
		return nil, err
	}
	account, err := accs.GetAccount(address)
	if err != nil {
		return nil, err
	}
	// If the account is unknown to us return a zeroed account
	if account == nil {
		account = &acm.Account{
			Address: address,
		}
	}
	pubKey, err := accs.keyClient.PublicKey(address)
	if err != nil {
		return nil, err
	}
	account.PublicKey = pubKey
	return &SigningAccount{
		Account: account,
		Signer:  signer,
	}, nil
}

func (accs *Accounts) SequentialSigningAccount(address crypto.Address) (*SequentialSigningAccount, error) {
	return &SequentialSigningAccount{
		Address:       address,
		accountLocker: accs.Mutex(address.Bytes()),
		getter: func() (*SigningAccount, error) {
			return accs.SigningAccount(address)
		},
	}, nil
}

type UnlockFunc func()

func (ssa *SequentialSigningAccount) Lock() (*SigningAccount, UnlockFunc, error) {
	ssa.accountLocker.Lock()
	account, err := ssa.getter()
	if err != nil {
		ssa.accountLocker.Unlock()
		return nil, nil, err
	}
	return account, ssa.accountLocker.Unlock, err
}
