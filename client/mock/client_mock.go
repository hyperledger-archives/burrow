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

package mock

import (
	"github.com/tendermint/go-crypto"

	acc "github.com/eris-ltd/eris-db/account"
	. "github.com/eris-ltd/eris-db/client"
	consensus_types "github.com/eris-ltd/eris-db/consensus/types"
	core_types "github.com/eris-ltd/eris-db/core/types"
	"github.com/eris-ltd/eris-db/txs"
	"github.com/eris-ltd/eris-db/logging/loggers"
)

// NOTE [ben] Compiler check to ensure ErisMockClient successfully implements
// eris-db/client.NodeClient
var _ NodeClient = (*MockNodeClient)(nil)

type MockNodeClient struct {
	accounts map[string]*acc.Account
}

func NewMockNodeClient() *MockNodeClient {
	return &MockNodeClient{
		accounts: make(map[string]*acc.Account),
	}
}

func (mock *MockNodeClient) Broadcast(transaction txs.Tx) (*txs.Receipt, error) {
	// make zero transaction receipt
	txReceipt := &txs.Receipt{
		TxHash:          make([]byte, 20, 20),
		CreatesContract: 0,
		ContractAddr:    make([]byte, 20, 20),
	}
	return txReceipt, nil
}

func (mock *MockNodeClient) DeriveWebsocketClient() (nodeWsClient NodeWebsocketClient, err error) {
	return nil, nil
}

func (mock *MockNodeClient) GetAccount(address []byte) (*acc.Account, error) {
	// make zero account
	var zero [32]byte
	copyAddressBytes := make([]byte, len(address), len(address))
	copy(copyAddressBytes, address)
	account := &acc.Account{
		Address:     copyAddressBytes,
		PubKey:      crypto.PubKey(crypto.PubKeyEd25519(zero)),
		Sequence:    0,
		Balance:     0,
		Code:        make([]byte, 0),
		StorageRoot: make([]byte, 0),
	}
	return account, nil
}

func (mock *MockNodeClient) MockAddAccount(account *acc.Account) {
	addressString := string(account.Address[:])
	mock.accounts[addressString] = account.Copy()
}

func (mock *MockNodeClient) Status() (ChainId []byte,
	ValidatorPublicKey []byte, LatestBlockHash []byte,
	BlockHeight int, LatestBlockTime int64, err error) {
	// make zero account
	var zero [32]byte
	ed25519 := crypto.PubKeyEd25519(zero)
	pub := crypto.PubKey(ed25519)

	// fill return values
	ChainId = make([]byte, 64)
	LatestBlockHash = make([]byte, 64)
	ValidatorPublicKey = pub.Bytes()
	BlockHeight = 0
	LatestBlockTime = 0
	return
}

// QueryContract executes the contract code at address with the given data
func (mock *MockNodeClient) QueryContract(callerAddress, calleeAddress, data []byte) (ret []byte, gasUsed int64, err error) {
	// return zero
	ret = make([]byte, 0)
	return ret, 0, nil
}

// QueryContractCode executes the contract code at address with the given data but with provided code
func (mock *MockNodeClient) QueryContractCode(address, code, data []byte) (ret []byte, gasUsed int64, err error) {
	// return zero
	ret = make([]byte, 0)
	return ret, 0, nil
}

func (mock *MockNodeClient) DumpStorage(address []byte) (storage *core_types.Storage, err error) {
	return nil, nil
}

func (mock *MockNodeClient) GetName(name string) (owner []byte, data string, expirationBlock int, err error) {
	return nil, "", 0, nil
}

func (mock *MockNodeClient) ListValidators() (blockHeight int, bondedValidators, unbondingValidators []consensus_types.Validator, err error) {
	return 0, nil, nil, nil
}

func (mock *MockNodeClient) Logger() loggers.InfoTraceLogger {
	return loggers.NewNoopInfoTraceLogger()
}
