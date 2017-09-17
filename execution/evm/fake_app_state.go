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

package evm

import (
	"fmt"

	"bytes"

	acm "github.com/hyperledger/burrow/account"
	. "github.com/hyperledger/burrow/word"
	"github.com/hyperledger/burrow/execution/evm/sha3"
)

type FakeAppState struct {
	accounts map[acm.Address]*acm.ConcreteAccount
	storage  map[string]Word256
}

func (fas *FakeAppState) GetAccount(addr acm.Address) *acm.ConcreteAccount {
	account := fas.accounts[addr]
	return account
}

func (fas *FakeAppState) UpdateAccount(account *acm.ConcreteAccount) {
	fas.accounts[account.Address] = account
}

func (fas *FakeAppState) RemoveAccount(account *acm.ConcreteAccount) {
	_, ok := fas.accounts[account.Address]
	if !ok {
		panic(fmt.Sprintf("Invalid account addr: %s", account.Address))
	} else {
		// Remove account
		delete(fas.accounts, account.Address)
	}
}

func (fas *FakeAppState) CreateAccount(creator *acm.ConcreteAccount) *acm.ConcreteAccount {
	addr := createAddress(creator)
	account := fas.accounts[addr]
	if account == nil {
		return &acm.ConcreteAccount{
			Address:  addr,
			Balance:  0,
			Code:     nil,
			Sequence: 0,
		}
	} else {
		panic(fmt.Sprintf("Invalid account addr: %s", addr))
	}
}

func (fas *FakeAppState) GetStorage(addr acm.Address, key Word256) Word256 {
	_, ok := fas.accounts[addr]
	if !ok {
		panic(fmt.Sprintf("Invalid account addr: %s", addr))
	}

	value, ok := fas.storage[addr.String()+key.String()]
	if ok {
		return value
	} else {
		return Zero256
	}
}

func (fas *FakeAppState) SetStorage(addr acm.Address, key Word256, value Word256) {
	_, ok := fas.accounts[addr]
	if !ok {

		fmt.Println("\n\n", fas.accountsDump())
		panic(fmt.Sprintf("Invalid account addr: %s", addr))
	}

	fas.storage[addr.String()+key.String()] = value
}

func (fas *FakeAppState) accountsDump() string {
	buf := new(bytes.Buffer)
	fmt.Fprint(buf, "Dumping accounts...", "\n")
	for _, acc := range fas.accounts {
		fmt.Fprint(buf, acc.Address.String(), "\n")
	}
	return buf.String()
}

// Creates a 20 byte address and bumps the nonce.
func createAddress(creator *acm.ConcreteAccount) acm.Address {
	sequence := creator.Sequence
	creator.Sequence += 1
	temp := make([]byte, 32+8)
	copy(temp, creator.Address[:])
	PutInt64BE(temp[32:], sequence)
	return acm.MustAddressFromBytes(sha3.Sha3(temp)[:20])
}
