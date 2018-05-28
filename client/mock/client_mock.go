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
	"time"

	acm "github.com/hyperledger/burrow/account"
	. "github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/go-crypto"
)

var _ NodeClient = (*MockNodeClient)(nil)

type MockNodeClient struct {
	accounts map[string]*acm.ConcreteAccount
}

func NewMockNodeClient() *MockNodeClient {
	return &MockNodeClient{
		accounts: make(map[string]*acm.ConcreteAccount),
	}
}

func (mock *MockNodeClient) Broadcast(transaction txs.Tx) (*txs.Receipt, error) {
	// make zero transaction receipt
	txReceipt := &txs.Receipt{
		TxHash:          make([]byte, 20),
		CreatesContract: false,
	}
	return txReceipt, nil
}

func (mock *MockNodeClient) DeriveWebsocketClient() (nodeWsClient NodeWebsocketClient, err error) {
	return nil, nil
}

func (mock *MockNodeClient) GetAccount(address acm.Address) (acm.Account, error) {
	// make zero account
	return acm.FromAddressable(acm.GeneratePrivateAccountFromSecret("mock-node-client-account")), nil
}

func (mock *MockNodeClient) MockAddAccount(account *acm.ConcreteAccount) {
	addressString := string(account.Address[:])
	mock.accounts[addressString] = account.Copy()
}

func (mock *MockNodeClient) Status() (ChainId []byte, ValidatorPublicKey []byte, LatestBlockHash []byte,
	BlockHeight uint64, LatestBlockTime time.Time, err error) {
	// fill return values
	ChainId = make([]byte, 64)
	LatestBlockHash = make([]byte, 64)
	ValidatorPublicKey = crypto.PubKeyEd25519{}.Wrap().Bytes()
	BlockHeight = 0
	LatestBlockTime = time.Time{}
	return
}

// QueryContract executes the contract code at address with the given data
func (mock *MockNodeClient) QueryContract(callerAddress, calleeAddress acm.Address,
	data []byte) (ret []byte, gasUsed uint64, err error) {

	// return zero
	ret = make([]byte, 0)
	return
}

// QueryContractCode executes the contract code at address with the given data but with provided code
func (mock *MockNodeClient) QueryContractCode(address acm.Address, code,
	data []byte) (ret []byte, gasUsed uint64, err error) {
	// return zero
	ret = make([]byte, 0)
	return
}

func (mock *MockNodeClient) DumpStorage(address acm.Address) (storage *rpc.ResultDumpStorage, err error) {
	return
}

func (mock *MockNodeClient) GetName(name string) (owner acm.Address, data string, expirationBlock uint64, err error) {
	return
}

func (mock *MockNodeClient) ListValidators() (blockHeight uint64, validators []string, err error) {
	return
}

func (mock *MockNodeClient) Logger() *logging.Logger {
	return logging.NewNoopLogger()
}
