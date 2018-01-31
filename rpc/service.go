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

package rpc

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/query"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/version"
	tm_types "github.com/tendermint/tendermint/types"
)

// Magic! Should probably be configurable, but not shouldn't be so huge we
// end up DoSing ourselves.
const MaxBlockLookback = 100

// Base service that provides implementation for all underlying RPC methods
type Service interface {
	// Transact
	Transactor() execution.Transactor
	// List mempool transactions pass -1 for all unconfirmed transactions
	ListUnconfirmedTxs(maxTxs int) (*ResultListUnconfirmedTxs, error)
	// Events
	Subscribe(subscriptionId, eventId string, callback func(eventData event.AnyEventData)) error
	Unsubscribe(subscriptionId string) error
	// Status
	Status() (*ResultStatus, error)
	NetInfo() (*ResultNetInfo, error)
	// Accounts
	GetAccount(address acm.Address) (*ResultGetAccount, error)
	ListAccounts(predicate func(acm.Account) bool) (*ResultListAccounts, error)
	GetStorage(address acm.Address, key []byte) (*ResultGetStorage, error)
	DumpStorage(address acm.Address) (*ResultDumpStorage, error)
	// Blockchain
	Genesis() (*ResultGenesis, error)
	ChainId() (*ResultChainId, error)
	GetBlock(height uint64) (*ResultGetBlock, error)
	ListBlocks(minHeight, maxHeight uint64) (*ResultListBlocks, error)
	// Consensus
	ListValidators() (*ResultListValidators, error)
	DumpConsensusState() (*ResultDumpConsensusState, error)
	Peers() (*ResultPeers, error)
	// Names
	GetName(name string) (*ResultGetName, error)
	ListNames(predicate func(*execution.NameRegEntry) bool) (*ResultListNames, error)
	// Private keys and signing
	GeneratePrivateAccount() (*ResultGeneratePrivateAccount, error)
}

type service struct {
	state        acm.StateIterable
	eventEmitter event.Emitter
	nameReg      execution.NameRegIterable
	blockchain   bcm.Blockchain
	transactor   execution.Transactor
	nodeView     query.NodeView
	logger       logging_types.InfoTraceLogger
}

var _ Service = &service{}
var _ event.Subscribable = Service(nil)

func NewService(state acm.StateIterable, nameReg execution.NameRegIterable, eventEmitter event.Emitter,
	blockchain bcm.Blockchain, transactor execution.Transactor, nodeView query.NodeView,
	logger logging_types.InfoTraceLogger) *service {

	return &service{
		state:        state,
		nameReg:      nameReg,
		eventEmitter: eventEmitter,
		blockchain:   blockchain,
		transactor:   transactor,
		nodeView:     nodeView,
		logger:       logger.With(structure.ComponentKey, "Service"),
	}
}

// Transacting...

func (s *service) Transactor() execution.Transactor {
	return s.transactor
}

func (s *service) ListUnconfirmedTxs(maxTxs int) (*ResultListUnconfirmedTxs, error) {
	// Get all transactions for now
	transactions, err := s.nodeView.MempoolTransactions(maxTxs)
	if err != nil {
		return nil, err
	}
	wrappedTxs := make([]txs.Wrapper, len(transactions))
	for i, tx := range transactions {
		wrappedTxs[i] = txs.Wrap(tx)
	}
	return &ResultListUnconfirmedTxs{
		N:   len(transactions),
		Txs: wrappedTxs,
	}, nil
}

// All methods in this file return (Result*, error) which is the return
// signature assumed by go-rpc
func (s *service) Subscribe(subscriptionId, eventId string, callback func(event.AnyEventData)) error {
	logging.InfoMsg(s.logger, "Subscribing to event",
		"eventId", eventId, "subscriptionId", subscriptionId)
	return s.eventEmitter.Subscribe(subscriptionId, eventId, callback)
}

func (s *service) Unsubscribe(subscriptionId string) error {
	return s.eventEmitter.Unsubscribe(subscriptionId)
}

func (s *service) Status() (*ResultStatus, error) {
	tip := s.blockchain.Tip()
	latestHeight := tip.LastBlockHeight()
	var (
		latestBlockMeta *tm_types.BlockMeta
		latestBlockHash []byte
		latestBlockTime int64
	)
	if latestHeight != 0 {
		latestBlockMeta = s.nodeView.BlockStore().LoadBlockMeta(int(latestHeight))
		latestBlockHash = latestBlockMeta.Header.Hash()
		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
	}
	return &ResultStatus{
		NodeInfo:          s.nodeView.NodeInfo(),
		GenesisHash:       s.blockchain.GenesisHash(),
		PubKey:            s.nodeView.PrivValidatorPublicKey(),
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime,
		NodeVersion:       version.GetVersionString(),
	}, nil
}

func (s *service) ChainId() (*ResultChainId, error) {
	return &ResultChainId{
		ChainName:   s.blockchain.GenesisDoc().ChainName,
		ChainId:     s.blockchain.ChainID(),
		GenesisHash: s.blockchain.GenesisHash(),
	}, nil
}

func (s *service) Peers() (*ResultPeers, error) {
	peers := make([]*Peer, s.nodeView.Peers().Size())
	for i, peer := range s.nodeView.Peers().List() {
		peers[i] = &Peer{
			NodeInfo:   peer.NodeInfo(),
			IsOutbound: peer.IsOutbound(),
		}
	}
	return &ResultPeers{
		Peers: peers,
	}, nil
}

func (s *service) NetInfo() (*ResultNetInfo, error) {
	listening := s.nodeView.IsListening()
	listeners := []string{}
	for _, listener := range s.nodeView.Listeners() {
		listeners = append(listeners, listener.String())
	}
	peers, err := s.Peers()
	if err != nil {
		return nil, err
	}
	return &ResultNetInfo{
		Listening: listening,
		Listeners: listeners,
		Peers:     peers.Peers,
	}, nil
}

func (s *service) Genesis() (*ResultGenesis, error) {
	return &ResultGenesis{
		Genesis: s.blockchain.GenesisDoc(),
	}, nil
}

// Accounts
func (s *service) GetAccount(address acm.Address) (*ResultGetAccount, error) {
	acc, err := s.state.GetAccount(address)
	if err != nil {
		return nil, err
	}
	return &ResultGetAccount{Account: acm.AsConcreteAccount(acc)}, nil
}

func (s *service) ListAccounts(predicate func(acm.Account) bool) (*ResultListAccounts, error) {
	accounts := make([]*acm.ConcreteAccount, 0)
	s.state.IterateAccounts(func(account acm.Account) (stop bool) {
		if predicate(account) {
			accounts = append(accounts, acm.AsConcreteAccount(account))
		}
		return
	})

	return &ResultListAccounts{
		BlockHeight: s.blockchain.Tip().LastBlockHeight(),
		Accounts:    accounts,
	}, nil
}

func (s *service) GetStorage(address acm.Address, key []byte) (*ResultGetStorage, error) {
	account, err := s.state.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %s", address)
	}

	value, err := s.state.GetStorage(address, binary.LeftPadWord256(key))
	if err != nil {
		return nil, err
	}
	if value == binary.Zero256 {
		return &ResultGetStorage{Key: key, Value: nil}, nil
	}
	return &ResultGetStorage{Key: key, Value: value.UnpadLeft()}, nil
}

func (s *service) DumpStorage(address acm.Address) (*ResultDumpStorage, error) {
	account, err := s.state.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}
	var storageItems []StorageItem
	s.state.IterateStorage(address, func(key, value binary.Word256) (stop bool) {
		storageItems = append(storageItems, StorageItem{Key: key.UnpadLeft(), Value: value.UnpadLeft()})
		return
	})
	return &ResultDumpStorage{
		StorageRoot:  account.StorageRoot(),
		StorageItems: storageItems,
	}, nil
}

// Name registry
func (s *service) GetName(name string) (*ResultGetName, error) {
	entry := s.nameReg.GetNameRegEntry(name)
	if entry == nil {
		return nil, fmt.Errorf("name %s not found", name)
	}
	return &ResultGetName{Entry: entry}, nil
}

func (s *service) ListNames(predicate func(*execution.NameRegEntry) bool) (*ResultListNames, error) {
	var names []*execution.NameRegEntry
	s.nameReg.IterateNameRegEntries(func(entry *execution.NameRegEntry) (stop bool) {
		if predicate(entry) {
			names = append(names, entry)
		}
		return
	})
	return &ResultListNames{
		BlockHeight: s.blockchain.Tip().LastBlockHeight(),
		Names:       names,
	}, nil
}

func (s *service) GetBlock(height uint64) (*ResultGetBlock, error) {
	return &ResultGetBlock{
		Block:     s.nodeView.BlockStore().LoadBlock(int(height)),
		BlockMeta: s.nodeView.BlockStore().LoadBlockMeta(int(height)),
	}, nil
}

// Returns the current blockchain height and metadata for a range of blocks
// between minHeight and maxHeight. Only returns maxBlockLookback block metadata
// from the top of the range of blocks.
// Passing 0 for maxHeight sets the upper height of the range to the current
// blockchain height.
func (s *service) ListBlocks(minHeight, maxHeight uint64) (*ResultListBlocks, error) {
	latestHeight := s.blockchain.Tip().LastBlockHeight()

	if minHeight == 0 {
		minHeight = 1
	}
	if maxHeight == 0 || latestHeight < maxHeight {
		maxHeight = latestHeight
	}
	if maxHeight > minHeight && maxHeight-minHeight > MaxBlockLookback {
		minHeight = maxHeight - MaxBlockLookback
	}

	var blockMetas []*tm_types.BlockMeta
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := s.nodeView.BlockStore().LoadBlockMeta(int(height))
		blockMetas = append(blockMetas, blockMeta)
	}

	return &ResultListBlocks{
		LastHeight: latestHeight,
		BlockMetas: blockMetas,
	}, nil
}

func (s *service) ListValidators() (*ResultListValidators, error) {
	// TODO: when we reintroduce support for bonding and unbonding update this
	// to reflect the mutable bonding state
	validators := s.blockchain.Validators()
	concreteValidators := make([]*acm.ConcreteValidator, len(validators))
	for i, validator := range validators {
		concreteValidators[i] = acm.AsConcreteValidator(validator)
	}
	return &ResultListValidators{
		BlockHeight:         s.blockchain.Tip().LastBlockHeight(),
		BondedValidators:    concreteValidators,
		UnbondingValidators: nil,
	}, nil
}

func (s *service) DumpConsensusState() (*ResultDumpConsensusState, error) {
	peerRoundState, err := s.nodeView.PeerRoundStates()
	if err != nil {
		return nil, err
	}
	return &ResultDumpConsensusState{
		RoundState:      s.nodeView.RoundState(),
		PeerRoundStates: peerRoundState,
	}, nil
}

func (s *service) GeneratePrivateAccount() (*ResultGeneratePrivateAccount, error) {
	privateAccount, err := acm.GeneratePrivateAccount()
	if err != nil {
		return nil, err
	}
	return &ResultGeneratePrivateAccount{
		PrivAccount: acm.AsConcretePrivateAccount(privateAccount),
	}, nil
}
