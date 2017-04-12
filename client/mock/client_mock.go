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

package mock

import (
	"github.com/tendermint/go-crypto"

	acc "github.com/hyperledger/burrow/account"
	. "github.com/hyperledger/burrow/client"
	consensus_types "github.com/hyperledger/burrow/consensus/types"
	core_types "github.com/hyperledger/burrow/core/types"
	"github.com/hyperledger/burrow/logging/loggers"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
)

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

func (mock *MockNodeClient) Logger() logging_types.InfoTraceLogger {
	return loggers.NewNoopInfoTraceLogger()
}
