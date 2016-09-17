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
	net := &network{td}
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

func (pipe *MockPipe) Net() definitions.Net {
	return pipe.net
}

func (pipe *MockPipe) Transactor() definitions.Transactor {
	return pipe.transactor
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
	testData *td.TestData
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
	testData *td.TestData
}

func (evntr *eventer) Subscribe(subId, event string, callback func(txs.EventData)) error {
	return nil
}

func (evntr *eventer) Unsubscribe(subId string) error {
	return nil
}

// NameReg
type namereg struct {
	testData *td.TestData
}

func (nmreg *namereg) Entry(key string) (*core_types.NameRegEntry, error) {
	return nmreg.testData.GetNameRegEntry.Output, nil
}

func (nmreg *namereg) Entries(filters []*event.FilterData) (*core_types.ResultListNames, error) {
	return nmreg.testData.GetNameRegEntries.Output, nil
}

// Net
type network struct {
	testData *td.TestData
}

func (net *network) Info() (*core_types.NetworkInfo, error) {
	return net.testData.GetNetworkInfo.Output, nil
}

func (net *network) ClientVersion() (string, error) {
	return net.testData.GetClientVersion.Output.ClientVersion, nil
}

func (net *network) Moniker() (string, error) {
	return net.testData.GetMoniker.Output.Moniker, nil
}

func (net *network) Listening() (bool, error) {
	return net.testData.IsListening.Output.Listening, nil
}

func (net *network) Listeners() ([]string, error) {
	return net.testData.GetListeners.Output.Listeners, nil
}

func (net *network) Peers() ([]*core_types.Peer, error) {
	return net.testData.GetPeers.Output, nil
}

func (net *network) Peer(address string) (*core_types.Peer, error) {
	// return net.testData.GetPeer.Output, nil
	return nil, nil
}

// Txs
type transactor struct {
	testData *td.TestData
}

func (trans *transactor) Call(fromAddress, toAddress, data []byte) (*core_types.Call, error) {
	return trans.testData.Call.Output, nil
}

func (trans *transactor) CallCode(from, code, data []byte) (*core_types.Call, error) {
	return trans.testData.CallCode.Output, nil
}

func (trans *transactor) BroadcastTx(tx txs.Tx) (*txs.Receipt, error) {
	return nil, nil
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
