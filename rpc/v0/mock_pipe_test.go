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

package v0

import (
	"fmt"

	account "github.com/monax/burrow/account"
	core_types "github.com/monax/burrow/core/types"
	definitions "github.com/monax/burrow/definitions"
	event "github.com/monax/burrow/event"

	blockchain_types "github.com/monax/burrow/blockchain/types"
	consensus_types "github.com/monax/burrow/consensus/types"
	logging_types "github.com/monax/burrow/logging/types"
	manager_types "github.com/monax/burrow/manager/types"
	"github.com/monax/burrow/txs"

	"github.com/monax/burrow/logging/loggers"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-p2p"
	mintTypes "github.com/tendermint/tendermint/types"
)

// Base struct.
type MockPipe struct {
	testData        TestData
	accounts        definitions.Accounts
	blockchain      blockchain_types.Blockchain
	consensusEngine consensus_types.ConsensusEngine
	events          event.EventEmitter
	namereg         definitions.NameReg
	transactor      definitions.Transactor
	logger          logging_types.InfoTraceLogger
}

// Create a new mock tendermint pipe.
func NewMockPipe(td *TestData) definitions.Pipe {
	return &MockPipe{
		testData:        *td,
		accounts:        &accounts{td},
		blockchain:      &chain{td},
		consensusEngine: &consensusEngine{td},
		events:          &eventer{td},
		namereg:         &namereg{td},
		transactor:      &transactor{td},
		logger:          loggers.NewNoopInfoTraceLogger(),
	}
}

// Create a mock pipe with default mock data.
func NewDefaultMockPipe() definitions.Pipe {
	return NewMockPipe(LoadTestData())
}

func (pipe *MockPipe) Accounts() definitions.Accounts {
	return pipe.accounts
}

func (pipe *MockPipe) Blockchain() blockchain_types.Blockchain {
	return pipe.blockchain
}

func (pipe *MockPipe) Events() event.EventEmitter {
	return pipe.events
}

func (pipe *MockPipe) NameReg() definitions.NameReg {
	return pipe.namereg
}

func (pipe *MockPipe) Transactor() definitions.Transactor {
	return pipe.transactor
}

func (pipe *MockPipe) Logger() logging_types.InfoTraceLogger {
	return pipe.logger
}

func (pipe *MockPipe) GetApplication() manager_types.Application {
	// TODO: [ben] mock application
	return nil
}

func (pipe *MockPipe) SetConsensusEngine(_ consensus_types.ConsensusEngine) error {
	// TODO: [ben] mock consensus engine
	return nil
}

func (pipe *MockPipe) GetConsensusEngine() consensus_types.ConsensusEngine {
	return pipe.consensusEngine
}

func (pipe *MockPipe) SetBlockchain(_ blockchain_types.Blockchain) error {
	// TODO: [ben] mock consensus engine
	return nil
}

func (pipe *MockPipe) GetBlockchain() blockchain_types.Blockchain {
	return nil
}

func (pipe *MockPipe) GetTendermintPipe() (definitions.TendermintPipe, error) {
	return nil, fmt.Errorf("Tendermint pipe is not supported by mocked pipe.")
}

func (pipe *MockPipe) GenesisHash() []byte {
	return pipe.testData.GetGenesisHash.Output.Hash
}

// Components

// Accounts
type accounts struct {
	testData *TestData
}

func (acc *accounts) GenPrivAccount() (*account.PrivAccount, error) {
	return acc.testData.GenPrivAccount.Output, nil
}

func (acc *accounts) GenPrivAccountFromKey(key []byte) (*account.PrivAccount, error) {
	return acc.testData.GenPrivAccount.Output, nil
}

func (acc *accounts) Accounts([]*event.FilterData) (*core_types.AccountList, error) {
	return acc.testData.GetAccounts.Output, nil
}

func (acc *accounts) Account(address []byte) (*account.Account, error) {
	return acc.testData.GetAccount.Output, nil
}

func (acc *accounts) Storage(address []byte) (*core_types.Storage, error) {
	return acc.testData.GetStorage.Output, nil
}

func (acc *accounts) StorageAt(address, key []byte) (*core_types.StorageItem, error) {
	return acc.testData.GetStorageAt.Output, nil
}

// Blockchain
type chain struct {
	testData *TestData
}

func (this *chain) ChainId() string {
	return this.testData.GetChainId.Output.ChainId
}

func (this *chain) Height() int {
	return this.testData.GetLatestBlockHeight.Output.Height
}

func (this *chain) Block(height int) *mintTypes.Block {
	return this.testData.GetBlock.Output
}

func (this *chain) BlockMeta(height int) *mintTypes.BlockMeta {
	return &mintTypes.BlockMeta{}
}

// Consensus
type consensusEngine struct {
	testData *TestData
}

func (cons *consensusEngine) BroadcastTransaction(transaction []byte,
	callback func(*abci_types.Response)) error {
	return nil
}

func (cons *consensusEngine) IsListening() bool {
	return cons.testData.IsListening.Output.Listening
}

func (cons *consensusEngine) Listeners() []p2p.Listener {
	p2pListeners := make([]p2p.Listener, 0)

	for _, name := range cons.testData.GetListeners.Output.Listeners {
		p2pListeners = append(p2pListeners, p2p.NewDefaultListener("tcp", name, true))
	}

	return p2pListeners
}

func (cons *consensusEngine) NodeInfo() *p2p.NodeInfo {
	return &p2p.NodeInfo{
		Version: cons.testData.GetNetworkInfo.Output.ClientVersion,
		Moniker: cons.testData.GetNetworkInfo.Output.Moniker,
	}
}

func (cons *consensusEngine) Peers() []*consensus_types.Peer {
	return cons.testData.GetPeers.Output
}

func (cons *consensusEngine) PublicValidatorKey() crypto.PubKey {
	return crypto.PubKeyEd25519{
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
	}
}

func (cons *consensusEngine) Events() event.EventEmitter {
	return nil
}

func (cons *consensusEngine) ListUnconfirmedTxs(maxTxs int) ([]txs.Tx, error) {
	return cons.testData.GetUnconfirmedTxs.Output.Txs, nil
}

func (cons *consensusEngine) ListValidators() []consensus_types.Validator {
	return nil
}

func (cons *consensusEngine) ConsensusState() *consensus_types.ConsensusState {
	return &consensus_types.ConsensusState{}
}

func (cons *consensusEngine) PeerConsensusStates() map[string]string {
	return map[string]string{}
}

// Events
type eventer struct {
	testData *TestData
}

func (evntr *eventer) Subscribe(subId, event string, callback func(txs.EventData)) error {
	return nil
}

func (evntr *eventer) Unsubscribe(subId string) error {
	return nil
}

// NameReg
type namereg struct {
	testData *TestData
}

func (nmreg *namereg) Entry(key string) (*core_types.NameRegEntry, error) {
	return nmreg.testData.GetNameRegEntry.Output, nil
}

func (nmreg *namereg) Entries(filters []*event.FilterData) (*core_types.ResultListNames, error) {
	return nmreg.testData.GetNameRegEntries.Output, nil
}

// Txs
type transactor struct {
	testData *TestData
}

func (trans *transactor) Call(fromAddress, toAddress, data []byte) (*core_types.Call, error) {
	return trans.testData.Call.Output, nil
}

func (trans *transactor) CallCode(from, code, data []byte) (*core_types.Call, error) {
	return trans.testData.CallCode.Output, nil
}

func (trans *transactor) BroadcastTx(tx txs.Tx) (*txs.Receipt, error) {
	receipt := txs.GenerateReceipt(trans.testData.GetChainId.Output.ChainId, tx)
	return &receipt, nil
}

func (trans *transactor) Transact(privKey, address, data []byte, gasLimit, fee int64) (*txs.Receipt, error) {
	if address == nil || len(address) == 0 {
		return trans.testData.TransactCreate.Output, nil
	}
	return trans.testData.Transact.Output, nil
}

func (trans *transactor) TransactAndHold(privKey, address, data []byte, gasLimit, fee int64) (*txs.EventDataCall, error) {
	return nil, nil
}

func (trans *transactor) Send(privKey, toAddress []byte, amount int64) (*txs.Receipt, error) {
	return nil, nil
}

func (trans *transactor) SendAndHold(privKey, toAddress []byte, amount int64) (*txs.Receipt, error) {
	return nil, nil
}

func (trans *transactor) TransactNameReg(privKey []byte, name, data string, amount, fee int64) (*txs.Receipt, error) {
	return trans.testData.TransactNameReg.Output, nil
}

func (trans *transactor) SignTx(tx txs.Tx, privAccounts []*account.PrivAccount) (txs.Tx, error) {
	return nil, nil
}
