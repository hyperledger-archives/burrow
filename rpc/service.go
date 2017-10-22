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
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
	tm_types "github.com/tendermint/tendermint/types"
)

// Magic! Should probably be configurable, but not shouldn't be so huge we
// end up DoSing ourselves.
const MaxBlockLookback = 100

// Base service that provides implementation for all underlying RPC methods
type Service interface {
	// Transact
	BroadcastTx(tx txs.Tx) (*ResultBroadcastTx, error)
	// Events
	Subscribe(eventId string, callback func(eventData event.AnyEventData)) (*ResultSubscribe, error)
	Unsubscribe(subscriptionId string) (*ResultUnsubscribe, error)
	// Status
	Status() (*ResultStatus, error)
	NetInfo() (*ResultNetInfo, error)
	// Accounts
	GetAccount(address acm.Address) (*ResultGetAccount, error)
	ListAccounts() (*ResultListAccounts, error)
	GetStorage(address acm.Address, key []byte) (*ResultGetStorage, error)
	DumpStorage(address acm.Address) (*ResultDumpStorage, error)
	// Simulated call
	Call(fromAddress, toAddress acm.Address, data []byte) (*ResultCall, error)
	CallCode(fromAddress acm.Address, code, data []byte) (*ResultCall, error)
	// Blockchain
	Genesis() (*ResultGenesis, error)
	ChainId() (*ResultChainId, error)
	BlockchainInfo(minHeight, maxHeight uint64) (*ResultBlockchainInfo, error)
	GetBlock(height uint64) (*ResultGetBlock, error)
	// Consensus
	ListUnconfirmedTxs(maxTxs int) (*ResultListUnconfirmedTxs, error)
	ListValidators() (*ResultListValidators, error)
	DumpConsensusState() (*ResultDumpConsensusState, error)
	Peers() (*ResultPeers, error)
	// Names
	GetName(name string) (*ResultGetName, error)
	ListNames() (*ResultListNames, error)
	// Private keys and signing
	SignTx(tx txs.Tx, concretePrivateAccounts []*acm.ConcretePrivateAccount) (*ResultSignTx, error)
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

func NewService(state acm.StateIterable, eventEmitter event.Emitter, nameReg execution.NameRegIterable,
	blockchain bcm.Blockchain, transactor execution.Transactor, nodeView query.NodeView,
	logger logging_types.InfoTraceLogger) *service {

	return &service{
		state:        state,
		eventEmitter: eventEmitter,
		nameReg:      nameReg,
		blockchain:   blockchain,
		transactor:   transactor,
		nodeView:     nodeView,
		logger:       logger,
	}
}

// All methods in this file return (Result*, error) which is the return
// signature assumed by go-rpc

func (s *service) Subscribe(eventId string, callback func(event.AnyEventData)) (*ResultSubscribe, error) {
	subscriptionId, err := event.GenerateSubId()

	logging.InfoMsg(s.logger, "Subscribing to event",
		"eventId", eventId, "subscriptionId", subscriptionId)
	if err != nil {
		return nil, err
	}
	err = s.eventEmitter.Subscribe(subscriptionId, eventId, callback)
	if err != nil {
		return nil, err
	}
	return &ResultSubscribe{
		SubscriptionId: subscriptionId,
		Event:          eventId,
	}, nil
}

func (s *service) Unsubscribe(subscriptionId string) (*ResultUnsubscribe, error) {
	err := s.eventEmitter.Unsubscribe(subscriptionId)
	if err != nil {
		return nil, err
	} else {
		return &ResultUnsubscribe{SubscriptionId: subscriptionId}, nil
	}
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
		PubKey:            s.nodeView.PrivValidatorPubKey(),
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime}, nil
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

func (s *service) ListAccounts() (*ResultListAccounts, error) {
	accounts := make([]*acm.ConcreteAccount, 0)
	s.state.IterateAccounts(func(account acm.Account) (stop bool) {
		accounts = append(accounts, acm.AsConcreteAccount(account))
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
	storageItems := []StorageItem{}
	s.state.IterateStorage(address, func(key, value binary.Word256) (stop bool) {
		storageItems = append(storageItems, StorageItem{Key: key.UnpadLeft(), Value: value.UnpadLeft()})
		return
	})
	return &ResultDumpStorage{
		StorageRoot:  account.StorageRoot(),
		StorageItems: storageItems,
	}, nil
}

func (s *service) Call(fromAddress, toAddress acm.Address, data []byte) (*ResultCall, error) {
	call, err := s.transactor.Call(fromAddress, toAddress, data)
	if err != nil {
		return nil, err
	}
	return &ResultCall{
		Call: call,
	}, nil
}

func (s *service) CallCode(fromAddress acm.Address, code, data []byte) (*ResultCall, error) {
	call, err := s.transactor.CallCode(fromAddress, code, data)
	if err != nil {
		return nil, err
	}
	return &ResultCall{
		Call: call,
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

func (s *service) ListNames() (*ResultListNames, error) {
	var names []*execution.NameRegEntry
	s.nameReg.IterateNameRegEntries(func(entry *execution.NameRegEntry) (stop bool) {
		names = append(names, entry)
		return false
	})
	return &ResultListNames{
		BlockHeight: s.blockchain.Tip().LastBlockHeight(),
		Names:       names,
	}, nil
}

func (s *service) BroadcastTx(tx txs.Tx) (*ResultBroadcastTx, error) {
	receipt, err := s.transactor.BroadcastTx(tx)
	if err != nil {
		return nil, err
	}
	return &ResultBroadcastTx{
		Receipt: receipt,
	}, nil
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

// Returns the current blockchain height and metadata for a range of blocks
// between minHeight and maxHeight. Only returns maxBlockLookback block metadata
// from the top of the range of blocks.
// Passing 0 for maxHeight sets the upper height of the range to the current
// blockchain height.
func (s *service) BlockchainInfo(minHeight, maxHeight uint64) (*ResultBlockchainInfo, error) {
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

	blockMetas := []*tm_types.BlockMeta{}
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := s.nodeView.BlockStore().LoadBlockMeta(int(height))
		blockMetas = append(blockMetas, blockMeta)
	}

	return &ResultBlockchainInfo{
		LastHeight: latestHeight,
		BlockMetas: blockMetas,
	}, nil
}

func (s *service) GetBlock(height uint64) (*ResultGetBlock, error) {
	return &ResultGetBlock{
		Block:     s.nodeView.BlockStore().LoadBlock(int(height)),
		BlockMeta: s.nodeView.BlockStore().LoadBlockMeta(int(height)),
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

// TODO: Either deprecate this or ensure it can only happen over secure transport
func (s *service) SignTx(tx txs.Tx, concretePrivateAccounts []*acm.ConcretePrivateAccount) (*ResultSignTx, error) {
	privateAccounts := make([]acm.PrivateAccount, len(concretePrivateAccounts))
	for i, cpa := range concretePrivateAccounts {
		privateAccounts[i] = cpa.PrivateAccount()
	}

	tx, err := s.transactor.SignTx(tx, privateAccounts)
	return &ResultSignTx{Tx: txs.Wrap(tx)}, err
}

func (s *service) GeneratePrivateAccount() (*ResultGeneratePrivateAccount, error) {
	return &ResultGeneratePrivateAccount{
		PrivAccount: acm.GeneratePrivateAccount().ConcretePrivateAccount,
	}, nil
}
