package mock

import (
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
	ep "github.com/eris-ltd/eris-db/erisdb/pipe"
	td "github.com/eris-ltd/eris-db/test/testdata/testdata"
)

// Base struct.
type MockPipe struct {
	testOutput *td.Output
	accounts   ep.Accounts
	blockchain ep.Blockchain
	consensus  ep.Consensus
	events     ep.EventEmitter
	net        ep.Net
	transactor ep.Transactor
}

// Create a new mock tendermint pipe.
func NewMockPipe(td *td.TestData) ep.Pipe {
	testOutput := td.Output
	accounts := &accounts{testOutput}
	blockchain := &blockchain{testOutput}
	consensus := &consensus{testOutput}
	events := &events{testOutput}
	net := &net{testOutput}
	transactor := &transactor{testOutput}
	return &MockPipe{
		testOutput,
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
	return NewMockPipe(td.LoadTestData())
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
	testOutput *td.Output
}

func (this *accounts) GenPrivAccount() (*account.PrivAccount, error) {
	return this.testOutput.GenPrivAccount, nil
}

func (this *accounts) GenPrivAccountFromKey(key []byte) (*account.PrivAccount, error) {
	return this.testOutput.GenPrivAccount, nil
}

func (this *accounts) Accounts([]*ep.FilterData) (*ep.AccountList, error) {
	return this.testOutput.Accounts, nil
}

func (this *accounts) Account(address []byte) (*account.Account, error) {
	return this.testOutput.Account, nil
}

func (this *accounts) Storage(address []byte) (*ep.Storage, error) {
	return this.testOutput.Storage, nil
}

func (this *accounts) StorageAt(address, key []byte) (*ep.StorageItem, error) {
	return this.testOutput.StorageAt, nil
}

// Blockchain
type blockchain struct {
	testOutput *td.Output
}

func (this *blockchain) Info() (*ep.BlockchainInfo, error) {
	return this.testOutput.BlockchainInfo, nil
}

func (this *blockchain) ChainId() (string, error) {
	return this.testOutput.ChainId.ChainId, nil
}

func (this *blockchain) GenesisHash() ([]byte, error) {
	return this.testOutput.GenesisHash.Hash, nil
}

func (this *blockchain) LatestBlockHeight() (int, error) {
	return this.testOutput.LatestBlockHeight.Height, nil
}

func (this *blockchain) LatestBlock() (*types.Block, error) {
	return nil, nil
}

func (this *blockchain) Blocks([]*ep.FilterData) (*ep.Blocks, error) {
	return this.testOutput.Blocks, nil
}

func (this *blockchain) Block(height int) (*types.Block, error) {
	return this.testOutput.Block, nil
}

// Consensus
type consensus struct {
	testOutput *td.Output
}

func (this *consensus) State() (*ep.ConsensusState, error) {
	return this.testOutput.ConsensusState, nil
}

func (this *consensus) Validators() (*ep.ValidatorList, error) {
	return this.testOutput.Validators, nil
}

// Events
type events struct {
	testOutput *td.Output
}

func (this *events) Subscribe(subId, event string, callback func(interface{})) (bool, error) {
	return true, nil
}

func (this *events) Unsubscribe(subId string) (bool, error) {
	return true, nil
}

// Net
type net struct {
	testOutput *td.Output
}

func (this *net) Info() (*ep.NetworkInfo, error) {
	return this.testOutput.NetworkInfo, nil
}

func (this *net) ClientVersion() (string, error) {
	return this.testOutput.ClientVersion.ClientVersion, nil
}

func (this *net) Moniker() (string, error) {
	return this.testOutput.Moniker.Moniker, nil
}

func (this *net) Listening() (bool, error) {
	return this.testOutput.Listening.Listening, nil
}

func (this *net) Listeners() ([]string, error) {
	return this.testOutput.Listeners.Listeners, nil
}

func (this *net) Peers() ([]*ep.Peer, error) {
	return this.testOutput.Peers, nil
}

func (this *net) Peer(address string) (*ep.Peer, error) {
	return nil, nil
}

// Txs
type transactor struct {
	testOutput *td.Output
}

func (this *transactor) Call(address, data []byte) (*ep.Call, error) {
	return nil, nil
}

func (this *transactor) CallCode(code, data []byte) (*ep.Call, error) {
	return this.testOutput.CallCode, nil
}

func (this *transactor) BroadcastTx(tx types.Tx) (*ep.Receipt, error) {
	return nil, nil
}

func (this *transactor) UnconfirmedTxs() (*ep.UnconfirmedTxs, error) {
	return this.testOutput.UnconfirmedTxs, nil
}

func (this *transactor) Transact(privKey, address, data []byte, gasLimit, fee int64) (*ep.Receipt, error) {
	if address == nil || len(address) == 0 {
		return this.testOutput.TxCreateReceipt, nil
	}
	return this.testOutput.TxReceipt, nil
}

func (this *transactor) SignTx(tx types.Tx, privAccounts []*account.PrivAccount) (types.Tx, error) {
	return nil, nil
}
