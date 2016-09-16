package mock

import (
	"fmt"

	account "github.com/eris-ltd/eris-db/account"
	core_types "github.com/eris-ltd/eris-db/core/types"
	definitions "github.com/eris-ltd/eris-db/definitions"
	event "github.com/eris-ltd/eris-db/event"

	blockchain_types "github.com/eris-ltd/eris-db/blockchain/types"
	consensus_types "github.com/eris-ltd/eris-db/consensus/types"
	manager_types "github.com/eris-ltd/eris-db/manager/types"
	td "github.com/eris-ltd/eris-db/test/testdata/testdata"
	"github.com/eris-ltd/eris-db/txs"

	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-p2p"
	mintTypes "github.com/tendermint/tendermint/types"
	tmsp_types "github.com/tendermint/tmsp/types"
)

// Base struct.
type MockPipe struct {
	testData        *td.TestData
	accounts        definitions.Accounts
	blockchain      blockchain_types.Blockchain
	consensusEngine consensus_types.ConsensusEngine
	events          event.EventEmitter
	namereg         definitions.NameReg
	net             definitions.Net
	transactor      definitions.Transactor
}

// Create a new mock tendermint pipe.
func NewMockPipe(td *td.TestData) definitions.Pipe {
	accounts := &accounts{td}
	blockchain := &blockchain{td}
	consensusEngine := &consensusEngine{td}
	eventer := &eventer{td}
	namereg := &namereg{td}
	net := &net{td}
	transactor := &transactor{td}
	return &MockPipe{
		td,
		accounts,
		blockchain,
		consensusEngine,
		eventer,
		namereg,
		net,
		transactor,
	}
}

// Create a mock pipe with default mock data.
func NewDefaultMockPipe() definitions.Pipe {
	return NewMockPipe(td.LoadTestData())
}

func (this *MockPipe) Accounts() definitions.Accounts {
	return this.accounts
}

func (this *MockPipe) Blockchain() blockchain_types.Blockchain {
	return this.blockchain
}

func (this *MockPipe) Consensus() consensus_types.ConsensusEngine {
	return this.consensusEngine
}

func (this *MockPipe) Events() event.EventEmitter {
	return this.events
}

func (this *MockPipe) NameReg() definitions.NameReg {
	return this.namereg
}

func (this *MockPipe) Net() definitions.Net {
	return this.net
}

func (this *MockPipe) Transactor() definitions.Transactor {
	return this.transactor
}

func (this *MockPipe) GetApplication() manager_types.Application {
	// TODO: [ben] mock application
	return nil
}

func (this *MockPipe) SetConsensusEngine(_ consensus_types.ConsensusEngine) error {
	// TODO: [ben] mock consensus engine
	return nil
}

func (this *MockPipe) GetConsensusEngine() consensus_types.ConsensusEngine {
	return nil
}

func (this *MockPipe) SetBlockchain(_ blockchain_types.Blockchain) error {
	// TODO: [ben] mock consensus engine
	return nil
}

func (this *MockPipe) GetBlockchain() blockchain_types.Blockchain {
	return nil
}

func (this *MockPipe) GetTendermintPipe() (definitions.TendermintPipe, error) {
	return nil, fmt.Errorf("Tendermint pipe is not supported by mocked pipe.")
}

func (this *MockPipe) GenesisHash() []byte {
	return this.testData.GetGenesisHash.Output.Hash
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

func (this *accounts) Accounts([]*event.FilterData) (*core_types.AccountList, error) {
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

func (this *blockchain) ChainId() string {
	return this.testData.GetChainId.Output.ChainId
}

func (this *blockchain) Height() int {
	return this.testData.GetLatestBlockHeight.Output.Height
}

func (this *blockchain) Block(height int) *mintTypes.Block {
	return this.testData.GetBlock.Output
}

func (this *blockchain) BlockMeta(height int) *mintTypes.BlockMeta {
	return &mintTypes.BlockMeta{}
}

// Consensus
type consensusEngine struct {
	testData *td.TestData
}

func (cons *consensusEngine) BroadcastTransaction(transaction []byte,
	callback func(*tmsp_types.Response)) error {
	return nil
}

func (cons *consensusEngine) IsListening() bool {
	return true
}

func (cons *consensusEngine) Listeners() []p2p.Listener {
	return make([]p2p.Listener, 0)
}

func (cons *consensusEngine) NodeInfo() *p2p.NodeInfo {
	return &p2p.NodeInfo{}
}

func (cons *consensusEngine) Peers() []consensus_types.Peer {
	return make([]consensus_types.Peer, 0)
}

func (cons *consensusEngine) PublicValidatorKey() crypto.PubKey {
	return crypto.PubKeyEd25519{
		1,2,3,4,5,6,7,8,
		1,2,3,4,5,6,7,8,
		1,2,3,4,5,6,7,8,
		1,2,3,4,5,6,7,8,
	}
}

func (cons *consensusEngine) Events() event.EventEmitter {
	return nil
}

func (cons *consensusEngine) ListUnconfirmedTxs(maxTxs int) ([]txs.Tx, error) {
	return nil, nil

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
	testData *td.TestData
}

func (this *eventer) Subscribe(subId, event string, callback func(txs.EventData)) error {
	return nil
}

func (this *eventer) Unsubscribe(subId string) error {
	return nil
}

// NameReg
type namereg struct {
	testData *td.TestData
}

func (this *namereg) Entry(key string) (*core_types.NameRegEntry, error) {
	return this.testData.GetNameRegEntry.Output, nil
}

func (this *namereg) Entries(filters []*event.FilterData) (*core_types.ResultListNames, error) {
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

func (this *transactor) BroadcastTx(tx txs.Tx) (*txs.Receipt, error) {
	return nil, nil
}

func (this *transactor) UnconfirmedTxs() (*txs.UnconfirmedTxs, error) {
	return this.testData.GetUnconfirmedTxs.Output, nil
}

func (this *transactor) Transact(privKey, address, data []byte, gasLimit, fee int64) (*txs.Receipt, error) {
	if address == nil || len(address) == 0 {
		return this.testData.TransactCreate.Output, nil
	}
	return this.testData.Transact.Output, nil
}

func (this *transactor) TransactAndHold(privKey, address, data []byte, gasLimit, fee int64) (*txs.EventDataCall, error) {
	return nil, nil
}

func (this *transactor) Send(privKey, toAddress []byte, amount int64) (*txs.Receipt, error) {
	return nil, nil
}

func (this *transactor) SendAndHold(privKey, toAddress []byte, amount int64) (*txs.Receipt, error) {
	return nil, nil
}

func (this *transactor) TransactNameReg(privKey []byte, name, data string, amount, fee int64) (*txs.Receipt, error) {
	return this.testData.TransactNameReg.Output, nil
}

func (this *transactor) SignTx(tx txs.Tx, privAccounts []*account.PrivAccount) (txs.Tx, error) {
	return nil, nil
}
