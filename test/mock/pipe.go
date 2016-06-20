package mock

import (
	"github.com/eris-ltd/eris-db/account"
	core_types "github.com/eris-ltd/eris-db/core/types"
	td "github.com/eris-ltd/eris-db/test/testdata/testdata"
	types "github.com/eris-ltd/eris-db/txs"

	mintTypes "github.com/tendermint/tendermint/types"
)

// Base struct.
type MockPipe struct {
	testData   *td.TestData
	accounts   core_types.Accounts
	blockchain core_types.Blockchain
	consensus  core_types.Consensus
	events     core_types.EventEmitter
	namereg    core_types.NameReg
	net        core_types.Net
	transactor core_types.Transactor
}

// Create a new mock tendermint pipe.
func NewMockPipe(td *td.TestData) core_types.Pipe {
	accounts := &accounts{td}
	blockchain := &blockchain{td}
	consensus := &consensus{td}
	eventer := &eventer{td}
	namereg := &namereg{td}
	net := &net{td}
	transactor := &transactor{td}
	return &MockPipe{
		td,
		accounts,
		blockchain,
		consensus,
		eventer,
		namereg,
		net,
		transactor,
	}
}

// Create a mock pipe with default mock data.
func NewDefaultMockPipe() core_types.Pipe {
	return NewMockPipe(td.LoadTestData())
}

func (this *MockPipe) Accounts() core_types.Accounts {
	return this.accounts
}

func (this *MockPipe) Blockchain() core_types.Blockchain {
	return this.blockchain
}

func (this *MockPipe) Consensus() core_types.Consensus {
	return this.consensus
}

func (this *MockPipe) Events() core_types.EventEmitter {
	return this.events
}

func (this *MockPipe) NameReg() core_types.NameReg {
	return this.namereg
}

func (this *MockPipe) Net() core_types.Net {
	return this.net
}

func (this *MockPipe) Transactor() core_types.Transactor {
	return this.transactor
}

// Components

// Accounts
type accounts struct {
	testData *td.TestData
}

func (this *accounts) GenPrivAccount() (*account.PrivAccount, error) {
	return this.testData.GenPrivAccount.Output, nil
}

func (this *accounts) GenPrivAccountFromKey(key []byte) (*account.PrivAccount, error) {
	return this.testData.GenPrivAccount.Output, nil
}

func (this *accounts) Accounts([]*core_types.FilterData) (*core_types.AccountList, error) {
	return this.testData.GetAccounts.Output, nil
}

func (this *accounts) Account(address []byte) (*account.Account, error) {
	return this.testData.GetAccount.Output, nil
}

func (this *accounts) Storage(address []byte) (*core_types.Storage, error) {
	return this.testData.GetStorage.Output, nil
}

func (this *accounts) StorageAt(address, key []byte) (*core_types.StorageItem, error) {
	return this.testData.GetStorageAt.Output, nil
}

// Blockchain
type blockchain struct {
	testData *td.TestData
}

func (this *blockchain) Info() (*core_types.BlockchainInfo, error) {
	return this.testData.GetBlockchainInfo.Output, nil
}

func (this *blockchain) ChainId() (string, error) {
	return this.testData.GetChainId.Output.ChainId, nil
}

func (this *blockchain) GenesisHash() ([]byte, error) {
	return this.testData.GetGenesisHash.Output.Hash, nil
}

func (this *blockchain) LatestBlockHeight() (int, error) {
	return this.testData.GetLatestBlockHeight.Output.Height, nil
}

func (this *blockchain) LatestBlock() (*mintTypes.Block, error) {
	return this.testData.GetLatestBlock.Output, nil
}

func (this *blockchain) Blocks([]*core_types.FilterData) (*core_types.Blocks, error) {
	return this.testData.GetBlocks.Output, nil
}

func (this *blockchain) Block(height int) (*mintTypes.Block, error) {
	return this.testData.GetBlock.Output, nil
}

// Consensus
type consensus struct {
	testData *td.TestData
}

func (this *consensus) State() (*core_types.ConsensusState, error) {
	return this.testData.GetConsensusState.Output, nil
}

func (this *consensus) Validators() (*core_types.ValidatorList, error) {
	return this.testData.GetValidators.Output, nil
}

// Events
type eventer struct {
	testData *td.TestData
}

func (this *eventer) Subscribe(subId, event string, callback func(events.EventData)) (bool, error) {
	return true, nil
}

func (this *eventer) Unsubscribe(subId string) (bool, error) {
	return true, nil
}

// NameReg
type namereg struct {
	testData *td.TestData
}

func (this *namereg) Entry(key string) (*types.NameRegEntry, error) {
	return this.testData.GetNameRegEntry.Output, nil
}

func (this *namereg) Entries(filters []*core_types.FilterData) (*core_types.ResultListNames, error) {
	return this.testData.GetNameRegEntries.Output, nil
}

// Net
type net struct {
	testData *td.TestData
}

func (this *net) Info() (*core_types.NetworkInfo, error) {
	return this.testData.GetNetworkInfo.Output, nil
}

func (this *net) ClientVersion() (string, error) {
	return this.testData.GetClientVersion.Output.ClientVersion, nil
}

func (this *net) Moniker() (string, error) {
	return this.testData.GetMoniker.Output.Moniker, nil
}

func (this *net) Listening() (bool, error) {
	return this.testData.IsListening.Output.Listening, nil
}

func (this *net) Listeners() ([]string, error) {
	return this.testData.GetListeners.Output.Listeners, nil
}

func (this *net) Peers() ([]*core_types.Peer, error) {
	return this.testData.GetPeers.Output, nil
}

func (this *net) Peer(address string) (*core_types.Peer, error) {
	// return this.testData.GetPeer.Output, nil
	return nil, nil
}

// Txs
type transactor struct {
	testData *td.TestData
}

func (this *transactor) Call(fromAddress, toAddress, data []byte) (*core_types.Call, error) {
	return this.testData.Call.Output, nil
}

func (this *transactor) CallCode(from, code, data []byte) (*core_types.Call, error) {
	return this.testData.CallCode.Output, nil
}

func (this *transactor) BroadcastTx(tx types.Tx) (*core_types.Receipt, error) {
	return nil, nil
}

func (this *transactor) UnconfirmedTxs() (*core_types.UnconfirmedTxs, error) {
	return this.testData.GetUnconfirmedTxs.Output, nil
}

func (this *transactor) Transact(privKey, address, data []byte, gasLimit, fee int64) (*core_types.Receipt, error) {
	if address == nil || len(address) == 0 {
		return this.testData.TransactCreate.Output, nil
	}
	return this.testData.Transact.Output, nil
}

func (this *transactor) TransactAndHold(privKey, address, data []byte, gasLimit, fee int64) (*types.EventDataCall, error) {
	return nil, nil
}

func (this *transactor) Send(privKey, toAddress []byte, amount int64) (*core_types.Receipt, error) {
	return nil, nil
}

func (this *transactor) SendAndHold(privKey, toAddress []byte, amount int64) (*core_types.Receipt, error) {
	return nil, nil
}

func (this *transactor) TransactNameReg(privKey []byte, name, data string, amount, fee int64) (*core_types.Receipt, error) {
	return this.testData.TransactNameReg.Output, nil
}

func (this *transactor) SignTx(tx types.Tx, privAccounts []*account.PrivAccount) (types.Tx, error) {
	return nil, nil
}
