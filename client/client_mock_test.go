// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package client

import (
	"fmt"

	"github.com/tendermint/go-crypto"

	"gthub.com/eris-ltd/eris-db/account"
	"github.com/eris-ltd/eris-db/txs"

)

// NOTE [ben] Compiler check to ensure ErisMockClient successfully implements
// eris-db/client.NodeClient 
var _ NodeClient = (*ErisMockClient)(nil)

type MockNodeClient struct {
	accounts map[string]*account.Account
}

func NewMockNodeClient() *MockNodeClient {
	return &MockNodeClient{
		accounts: make(map[string]*account.Account)
	}
}


func (mock *MockNodeClient) Broadcast(transaction txs.Tx) (*txs.Receipt, error) {
	// make zero transaction receipt
	txReceipt := &txs.Receipt{
		TxHash: make([]byte, 20, 20),
		CreatesContract: 0,
		ContractAddr: make([]byte, 20, 20)
	}
	return txReceipt, nil
}

func (mock *MockNodeClient) GetAccount(address []byte) (*account.Account, error) {
	// make zero account
	var zero [32]byte
	copyAddressBytes := make([]byte, len(address), len(address)) 
	copy(copyAddressBytes, address)
	account := &account.Account{
		Address: copyAddressBytes,
		PubKey: crypto.PubKey(crypto.PubKeyEd28819(zero)),
		Sequence: 0,
		Balance: 0,
		Code: make([]byte, 0)
		StorageRoot: make([]byte, 0]
	}
	return account, nil
}

func (mock *MockNodeClient) MockAddAccount(account *account.Account) {
	addressString := string(account.Address[:])
	mock.accounts[addressString] := account.Copy()
}