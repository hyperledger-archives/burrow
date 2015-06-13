package mock

import (
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	"github.com/tendermint/tendermint/account"
	"github.com/tendermint/tendermint/types"
)

// Base struct.
type MockPipe struct {
	mockData   *MockData
	accounts   ep.Accounts
	blockchain ep.Blockchain
	consensus  ep.Consensus
	events     ep.EventEmitter
	net        ep.Net
	transactor ep.Transactor
}

// Create a new mock tendermint pipe.
func NewMockPipe(mockData *MockData) ep.Pipe {
	accounts := &accounts{mockData}
	blockchain := &blockchain{mockData}
	consensus := &consensus{mockData}
	events := &events{mockData}
	net := &net{mockData}
	transactor := &transactor{mockData}
	return &MockPipe{
		mockData,
		accounts,
		blockchain,
		consensus,
		events,
		net,
		transactor,
	}
}

// Create a mock pipe with default mock data.
func NewDefaultMockPipe() ep.Pipe {
	return NewMockPipe(NewDefaultMockData())
}

func (this *MockPipe) Accounts() ep.Accounts {
	return this.accounts
}

func (this *MockPipe) Blockchain() ep.Blockchain {
	return this.blockchain
}

func (this *MockPipe) Consensus() ep.Consensus {
	return this.consensus
}

func (this *MockPipe) Events() ep.EventEmitter {
	return this.events
}

func (this *MockPipe) Net() ep.Net {
	return this.net
}

func (this *MockPipe) Transactor() ep.Transactor {
	return this.transactor
}

// Components

// Accounts
type accounts struct {
	mockData *MockData
}

func (this *accounts) GenPrivAccount() (*account.PrivAccount, error) {
	return this.mockData.PrivAccount, nil
}

func (this *accounts) GenPrivAccountFromKey(key []byte) (*account.PrivAccount, error) {
	return this.mockData.PrivAccount, nil
}

func (this *accounts) Accounts([]*ep.FilterData) (*ep.AccountList, error) {
	return this.mockData.Accounts, nil
}

func (this *accounts) Account(address []byte) (*account.Account, error) {
	return this.mockData.Account, nil
}

func (this *accounts) Storage(address []byte) (*ep.Storage, error) {
	return this.mockData.Storage, nil
}

func (this *accounts) StorageAt(address, key []byte) (*ep.StorageItem, error) {
	return this.mockData.StorageAt, nil
}

// Blockchain
type blockchain struct {
	mockData *MockData
}

func (this *blockchain) Info() (*ep.BlockchainInfo, error) {
	return this.mockData.BlockchainInfo, nil
}

func (this *blockchain) ChainId() (string, error) {
	return this.mockData.ChainId.ChainId, nil
}

func (this *blockchain) GenesisHash() ([]byte, error) {
	return this.mockData.GenesisHash.Hash, nil
}

func (this *blockchain) LatestBlockHeight() (uint, error) {
	return this.mockData.LatestBlockHeight.Height, nil
}

func (this *blockchain) LatestBlock() (*types.Block, error) {
	return this.mockData.LatestBlock, nil
}

func (this *blockchain) Blocks([]*ep.FilterData) (*ep.Blocks, error) {
	return this.mockData.Blocks, nil
}

func (this *blockchain) Block(height uint) (*types.Block, error) {
	return this.mockData.Block, nil
}

// Consensus
type consensus struct {
	mockData *MockData
}

func (this *consensus) State() (*ep.ConsensusState, error) {
	return this.mockData.ConsensusState, nil
}

func (this *consensus) Validators() (*ep.ValidatorList, error) {
	return this.mockData.Validators, nil
}

// Events
type events struct {
	mockData *MockData
}

func (this *events) Subscribe(subId, event string, callback func(interface{})) (bool, error) {
	return true, nil
}

func (this *events) Unsubscribe(subId string) (bool, error) {
	return this.mockData.EventUnSub.Result, nil
}

// Net
type net struct {
	mockData *MockData
}

func (this *net) Info() (*ep.NetworkInfo, error) {
	return this.mockData.NetworkInfo, nil
}

func (this *net) Moniker() (string, error) {
	return this.mockData.Moniker.Moniker, nil
}

func (this *net) Listening() (bool, error) {
	return this.mockData.Listening.Listening, nil
}

func (this *net) Listeners() ([]string, error) {
	return this.mockData.Listeners.Listeners, nil
}

func (this *net) Peers() ([]*ep.Peer, error) {
	return this.mockData.Peers, nil
}

func (this *net) Peer(address string) (*ep.Peer, error) {
	return this.mockData.Peer, nil
}

// Txs
type transactor struct {
	mockData *MockData
}

func (this *transactor) Call(address, data []byte) (*ep.Call, error) {
	return this.mockData.Call, nil
}

func (this *transactor) CallCode(code, data []byte) (*ep.Call, error) {
	return this.mockData.CallCode, nil
}

func (this *transactor) BroadcastTx(tx types.Tx) (*ep.Receipt, error) {
	return this.mockData.BroadcastTx, nil
}

func (this *transactor) UnconfirmedTxs() (*ep.UnconfirmedTxs, error) {
	return this.mockData.UnconfirmedTxs, nil
}

func (this *transactor) TransactAsync(privKey, address, data []byte, gasLimit, fee uint64) (*ep.TransactionResult, error) {
	return nil, nil
}

func (this *transactor) Transact(privKey, address, data []byte, gasLimit, fee uint64) (*ep.Receipt, error) {
	return this.mockData.BroadcastTx, nil
}

func (this *transactor) SignTx(tx types.Tx, privAccounts []*account.PrivAccount) (types.Tx, error) {
	return this.mockData.SignTx, nil
}
