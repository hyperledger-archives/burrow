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

package core

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/query"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/word"
	"github.com/tendermint/tendermint/rpc/lib/types"
	tm_types "github.com/tendermint/tendermint/types"
)

// Magic! Should probably be configurable, but not shouldn't be so huge we
// end up DoSing ourselves.
const MaxBlockLookback = 100

// Base service that provides implementation for all underlying RPC methods
type Service interface {
	Subscribe(wsCtx rpctypes.WSRPCContext, eventId string) (*ResultSubscribe, error)
	Unsubscribe(wsCtx rpctypes.WSRPCContext, subscriptionId string) (*ResultUnsubscribe, error)
	Status() (*ResultStatus, error)
	ChainId() (*ResultChainId, error)
	Peers() ([]*Peer, error)
	NetInfo() (*ResultNetInfo, error)
	Genesis() (*ResultGenesis, error)
	GetAccount(addressBytes []byte) (*ResultGetAccount, error)
	//ListAccounts() (*ResultListAccounts, error)
	ListAccounts() (BurrowResult, error)
	GetStorage(addressBytes, key []byte) (*ResultGetStorage, error)
	DumpStorage(addressBytes []byte) (*ResultDumpStorage, error)
	Call(fromAddressBytes, toAddressBytes, data []byte) (*execution.Call, error)
	CallCode(fromAddressBytes, code, data []byte) (*execution.Call, error)
	GetName(name string) (*ResultGetName, error)
	ListNames() (*ResultListNames, error)
	BroadcastTx(tx txs.Tx) (*ResultBroadcastTx, error)
	ListUnconfirmedTxs(maxTxs int) (*ResultListUnconfirmedTxs, error)
	BlockchainInfo(minHeight, maxHeight uint64) (*ResultBlockchainInfo, error)
	GetBlock(height uint64) (*ResultGetBlock, error)
	ListValidators() (*ResultListValidators, error)
	DumpConsensusState() (*ResultDumpConsensusState, error)
	SignTx(tx txs.Tx, concretePrivateAccounts []*acm.ConcretePrivateAccount) (*ResultSignTx, error)
	GeneratePrivateAccount() (*ResultGeneratePrivateAccount, error)
}

type service struct {
	state        acm.StateIterable
	eventEmitter event.EventEmitter
	nameReg      execution.NameRegIterable
	blockchain   bcm.Blockchain
	transactor   execution.Transactor
	nodeView     query.NodeView
	logger       logging_types.InfoTraceLogger
}

var _ Service = &service{}

func NewService(state acm.StateIterable, eventEmitter event.EventEmitter, nameReg execution.NameRegIterable,
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

func (s *service) Subscribe(wsCtx rpctypes.WSRPCContext, eventId string) (*ResultSubscribe, error) {
	// NOTE: RPCResponses of subscribed events have id suffix "#event"
	// TODO: we really ought to allow multiple subscriptions from the same client address
	// to the same event. The code as it stands reflects the somewhat broken tendermint
	// implementation. We can use GenerateSubId to randomize the subscriptions id
	// and return it in the result. This would require clients to hang on to a
	// subscription id if they wish to unsubscribe, but then again they can just
	// drop their connection
	subscriptionId, err := event.GenerateSubId()
	if err != nil {
		return nil, err
		logging.InfoMsg(s.logger, "Subscribing to event",
			"eventId", eventId, "subscriptionId", subscriptionId)
	}
	s.eventEmitter.Subscribe(subscriptionId, eventId,
		func(eventData evm.EventData) {
			result := BurrowResult(
				&ResultEvent{
					Event: eventId,
					Data:  evm.EventData(eventData)})
			// NOTE: EventSwitch callbacks must be nonblocking
			wsCtx.GetRemoteAddr()
			// NOTE: EventSwitch callbacks must be nonblocking
			wsCtx.TryWriteRPCResponse(rpctypes.NewRPCSuccessResponse(wsCtx.Request.ID+"#event", &result))
		})
	return &ResultSubscribe{
		SubscriptionId: subscriptionId,
		Event:          eventId,
	}, nil
}

func (s *service) Unsubscribe(wsCtx rpctypes.WSRPCContext, subscriptionId string) (*ResultUnsubscribe, error) {
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
	chainId := s.blockchain.ChainID()

	return &ResultChainId{
		ChainName:   chainId, // TODO: remove ChainName, we should stick with Tendermint's human readable notion of ChainID
		ChainId:     chainId,
		GenesisHash: s.blockchain.GenesisHash(),
	}, nil
}

func (s *service) Peers() ([]*Peer, error) {
	peers := make([]*Peer, s.nodeView.Peers().Size())
	for i, peer := range s.nodeView.Peers().List() {
		peers[i] = &Peer{
			NodeInfo:   peer.NodeInfo(),
			IsOutbound: peer.IsOutbound(),
		}
	}
	return peers, nil
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
		Peers:     peers,
	}, nil
}

func (s *service) Genesis() (*ResultGenesis, error) {
	return &ResultGenesis{
		Genesis: s.blockchain.GenesisDoc(),
	}, nil
}

// Accounts
func (s *service) GetAccount(addressBytes []byte) (*ResultGetAccount, error) {
	address, err := acm.AddressFromBytes(addressBytes)
	if err != nil {
		return nil, err
	}

	acc, err := s.state.GetAccount(address)
	if err != nil {
		return nil, err
	}
	return &ResultGetAccount{Account: acc}, nil
}

func (s *service) ListAccounts() (BurrowResult, error) {
	accounts := make([]acm.Account, 0)
	s.state.IterateAccounts(func(account acm.Account) (stop bool) {
		accounts = append(accounts, account)
		return
	})

	return &ResultListAccounts{
		BlockHeight: s.blockchain.Tip().LastBlockHeight(),
		Accounts:    accounts,
	}, nil
}

func (s *service) GetStorage(addressBytes, key []byte) (*ResultGetStorage, error) {
	address, err := acm.AddressFromBytes(addressBytes)
	if err != nil {
		return nil, fmt.Errorf("GetStorage could not get address: %v", err)
	}
	account, err := s.state.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", addressBytes)
	}

	value, err := s.state.GetStorage(address, word.LeftPadWord256(key))
	if err != nil {
		return nil, err
	}
	if value == word.Zero256 {
		return &ResultGetStorage{Key: key, Value: nil}, nil
	}
	return &ResultGetStorage{Key: key, Value: value.UnpadLeft()}, nil
}

func (s *service) DumpStorage(addressBytes []byte) (*ResultDumpStorage, error) {
	address, err := acm.AddressFromBytes(addressBytes)
	if err != nil {
		return nil, fmt.Errorf("DumpStorage could not get address: %v", err)
	}
	account, err := s.state.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}
	storageItems := []StorageItem{}
	s.state.IterateStorage(address, func(key, value word.Word256) (stop bool) {
		storageItems = append(storageItems, StorageItem{Key: key.UnpadLeft(), Value: value.UnpadLeft()})
		return
	})
	return &ResultDumpStorage{
		StorageRoot:  account.StorageRoot(),
		StorageItems: storageItems,
	}, nil
}

func (s *service) Call(fromAddressBytes, toAddressBytes, data []byte) (*execution.Call, error) {
	fromAddress, err := acm.AddressFromBytes(fromAddressBytes)
	if err != nil {
		return nil, fmt.Errorf("Call could not get 'from' address: %v", err)
	}

	toAddress, err := acm.AddressFromBytes(toAddressBytes)
	if err != nil {
		return nil, fmt.Errorf("Call could not get 'to' address: %v", err)
	}

	return s.transactor.Call(fromAddress, toAddress, data)
}

func (s *service) CallCode(fromAddressBytes, code, data []byte) (*execution.Call, error) {
	fromAddress, err := acm.AddressFromBytes(fromAddressBytes)
	if err != nil {
		return nil, fmt.Errorf("CallCode could not get 'from' address: %v", err)
	}

	return s.transactor.CallCode(fromAddress, code, data)
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
		Receipt: *receipt,
	}, nil
}

func (s *service) ListUnconfirmedTxs(maxTxs int) (*ResultListUnconfirmedTxs, error) {
	// Get all transactions for now
	transactions, err := s.nodeView.MempoolTransactions(maxTxs)
	if err != nil {
		return nil, err
	}
	return &ResultListUnconfirmedTxs{
		N:   len(transactions),
		Txs: transactions,
	}, nil
}

// Returns the current blockchain height and metadata for a range of blocks
// between minHeight and maxHeight. Only returns maxBlockLookback block metadata
// from the top of the range of blocks.
// Passing 0 for maxHeight sets the upper height of the range to the current
// blockchain height.
func (s *service) BlockchainInfo(minHeight, maxHeight uint64) (*ResultBlockchainInfo, error) {
	latestHeight := s.blockchain.Tip().LastBlockHeight()

	if maxHeight < 1 || latestHeight < maxHeight {
		maxHeight = latestHeight
	}
	if minHeight < 1 {
		lookbackHeight := maxHeight - MaxBlockLookback
		if lookbackHeight > 1 {
			minHeight = lookbackHeight
		} else {
			minHeight = 1
		}
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
	return &ResultListValidators{
		BlockHeight:         s.blockchain.Tip().LastBlockHeight(),
		BondedValidators:    s.blockchain.Validators(),
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
	return &ResultSignTx{Tx: tx}, err
}

func (s *service) GeneratePrivateAccount() (*ResultGeneratePrivateAccount, error) {
	return &ResultGeneratePrivateAccount{
		PrivAccount: acm.GeneratePrivateAccount().ConcretePrivateAccount,
	}, nil
}
