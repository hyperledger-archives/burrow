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
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/project"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/rpc/core/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

// Magic! Should probably be configurable, but not shouldn't be so huge we
// end up DoSing ourselves.
const MaxBlockLookback = 1000

// Base service that provides implementation for all underlying RPC methods
type Service struct {
	state      state.IterableReader
	nameReg    names.IterableReader
	blockchain bcm.BlockchainInfo
	nodeView   *tendermint.NodeView
	logger     *logging.Logger
}

// Service provides an internal query and information service with serialisable return types on which can accomodate
// a number of transport front ends
func NewService(state state.IterableReader, nameReg names.IterableReader, blockchain bcm.BlockchainInfo,
	nodeView *tendermint.NodeView, logger *logging.Logger) *Service {

	return &Service{
		state:      state,
		nameReg:    nameReg,
		blockchain: blockchain,
		nodeView:   nodeView,
		logger:     logger.With(structure.ComponentKey, "Service"),
	}
}

func (s *Service) State() state.Reader {
	return s.state
}

func (s *Service) BlockchainInfo() bcm.BlockchainInfo {
	return s.blockchain
}

func (s *Service) ChainID() string {
	return s.blockchain.ChainID()
}

func (s *Service) UnconfirmedTxs(maxTxs int64) (*ResultUnconfirmedTxs, error) {
	// Get all transactions for now
	transactions, err := s.nodeView.MempoolTransactions(int(maxTxs))
	if err != nil {
		return nil, err
	}
	wrappedTxs := make([]*txs.Envelope, len(transactions))
	for i, tx := range transactions {
		wrappedTxs[i] = tx
	}
	return &ResultUnconfirmedTxs{
		NumTxs: len(transactions),
		Txs:    wrappedTxs,
	}, nil
}

func (s *Service) Status() (*ResultStatus, error) {
	return Status(s.BlockchainInfo(), s.nodeView, "", "")
}

func (s *Service) StatusWithin(blockTimeWithin, blockSeenTimeWithin string) (*ResultStatus, error) {
	return Status(s.BlockchainInfo(), s.nodeView, blockTimeWithin, blockSeenTimeWithin)
}

func (s *Service) ChainIdentifiers() (*ResultChainId, error) {
	return &ResultChainId{
		ChainName:   s.blockchain.GenesisDoc().ChainName,
		ChainId:     s.blockchain.ChainID(),
		GenesisHash: s.blockchain.GenesisHash(),
	}, nil
}

func (s *Service) Peers() []core_types.Peer {
	p2pPeers := s.nodeView.Peers().List()
	peers := make([]core_types.Peer, len(p2pPeers))
	for i, peer := range p2pPeers {
		peers[i] = core_types.Peer{
			NodeInfo:         peer.NodeInfo(),
			IsOutbound:       peer.IsOutbound(),
			ConnectionStatus: peer.Status(),
		}
	}
	return peers
}

func (s *Service) Network() (*ResultNetwork, error) {
	var listeners []string
	for _, listener := range s.nodeView.Listeners() {
		listeners = append(listeners, listener.String())
	}
	peers := s.Peers()
	return &ResultNetwork{
		ThisNode: s.nodeView.NodeInfo(),
		ResultNetInfo: &core_types.ResultNetInfo{
			Listening: s.nodeView.IsListening(),
			Listeners: listeners,
			NPeers:    len(peers),
			Peers:     peers,
		},
	}, nil
}

func (s *Service) Genesis() (*ResultGenesis, error) {
	return &ResultGenesis{
		Genesis: s.blockchain.GenesisDoc(),
	}, nil
}

// Accounts
func (s *Service) Account(address crypto.Address) (*ResultAccount, error) {
	acc, err := s.state.GetAccount(address)
	if err != nil {
		return nil, err
	}
	return &ResultAccount{Account: acm.AsConcreteAccount(acc)}, nil
}

func (s *Service) Accounts(predicate func(acm.Account) bool) (*ResultAccounts, error) {
	accounts := make([]*acm.ConcreteAccount, 0)
	s.state.IterateAccounts(func(account acm.Account) (stop bool) {
		if predicate(account) {
			accounts = append(accounts, acm.AsConcreteAccount(account))
		}
		return
	})

	return &ResultAccounts{
		BlockHeight: s.blockchain.LastBlockHeight(),
		Accounts:    accounts,
	}, nil
}

func (s *Service) Storage(address crypto.Address, key []byte) (*ResultStorage, error) {
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
		return &ResultStorage{Key: key, Value: nil}, nil
	}
	return &ResultStorage{Key: key, Value: value.UnpadLeft()}, nil
}

func (s *Service) DumpStorage(address crypto.Address) (*ResultDumpStorage, error) {
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
		StorageItems: storageItems,
	}, nil
}

func (s *Service) AccountHumanReadable(address crypto.Address) (*ResultAccountHumanReadable, error) {
	acc, err := s.state.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return &ResultAccountHumanReadable{}, nil
	}
	tokens, err := acc.Code().Tokens()
	if acc == nil {
		return &ResultAccountHumanReadable{}, nil
	}
	perms := permission.BasePermissionsToStringList(acc.Permissions().Base)
	if acc == nil {
		return &ResultAccountHumanReadable{}, nil
	}
	return &ResultAccountHumanReadable{
		Account: &AccountHumanReadable{
			Address:     acc.Address(),
			PublicKey:   acc.PublicKey(),
			Sequence:    acc.Sequence(),
			Balance:     acc.Balance(),
			Code:        tokens,
			Permissions: perms,
			Roles:       acc.Permissions().Roles,
		},
	}, nil
}

// Name registry
func (s *Service) Name(name string) (*ResultName, error) {
	entry, err := s.nameReg.GetName(name)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, fmt.Errorf("name %s not found", name)
	}
	return &ResultName{Entry: entry}, nil
}

func (s *Service) Names(predicate func(*names.Entry) bool) (*ResultNames, error) {
	var nms []*names.Entry
	s.nameReg.IterateNames(func(entry *names.Entry) (stop bool) {
		if predicate(entry) {
			nms = append(nms, entry)
		}
		return
	})
	return &ResultNames{
		BlockHeight: s.blockchain.LastBlockHeight(),
		Names:       nms,
	}, nil
}

func (s *Service) Block(height uint64) (*ResultBlock, error) {
	return &ResultBlock{
		Block:     &Block{s.nodeView.BlockStore().LoadBlock(int64(height))},
		BlockMeta: &BlockMeta{s.nodeView.BlockStore().LoadBlockMeta(int64(height))},
	}, nil
}

// Returns the current blockchain height and metadata for a range of blocks
// between minHeight and maxHeight. Only returns maxBlockLookback block metadata
// from the top of the range of blocks.
// Passing 0 for maxHeight sets the upper height of the range to the current
// blockchain height.
func (s *Service) Blocks(minHeight, maxHeight int64) (*ResultBlocks, error) {
	latestHeight := int64(s.blockchain.LastBlockHeight())

	if minHeight < 1 {
		minHeight = latestHeight
	}
	if maxHeight == 0 || latestHeight < maxHeight {
		maxHeight = latestHeight
	}
	if maxHeight > minHeight && maxHeight-minHeight > MaxBlockLookback {
		minHeight = maxHeight - MaxBlockLookback
	}

	var blockMetas []*tmTypes.BlockMeta
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := s.nodeView.BlockStore().LoadBlockMeta(height)
		blockMetas = append(blockMetas, blockMeta)
	}

	return &ResultBlocks{
		LastHeight: uint64(latestHeight),
		BlockMetas: blockMetas,
	}, nil
}

func (s *Service) Validators() (*ResultValidators, error) {
	validators := make([]*validator.Validator, 0, s.blockchain.NumValidators())
	s.blockchain.Validators().Iterate(func(id crypto.Addressable, power *big.Int) (stop bool) {
		address := id.Address()
		validators = append(validators, &validator.Validator{
			Address:   &address,
			PublicKey: id.PublicKey(),
			Power:     power.Uint64(),
		})
		return
	})
	return &ResultValidators{
		BlockHeight:         s.blockchain.LastBlockHeight(),
		BondedValidators:    validators,
		UnbondingValidators: nil,
	}, nil
}

func (s *Service) ConsensusState() (*ResultConsensusState, error) {
	peers := s.nodeView.Peers().List()
	peerStates := make([]core_types.PeerStateInfo, len(peers))
	for i, peer := range peers {
		peerState := peer.Get(tmTypes.PeerStateKey).(*consensus.PeerState)
		peerStateJSON, err := peerState.ToJSON()
		if err != nil {
			return nil, err
		}
		peerStates[i] = core_types.PeerStateInfo{
			// Peer basic info.
			NodeAddress: p2p.IDAddressString(peer.ID(), peer.NodeInfo().ListenAddr),
			// Peer consensus state.
			PeerState: peerStateJSON,
		}
	}

	roundStateJSON, err := s.nodeView.RoundStateJSON()
	if err != nil {
		return nil, err
	}
	return &ResultConsensusState{
		ResultDumpConsensusState: &core_types.ResultDumpConsensusState{
			RoundState: roundStateJSON,
			Peers:      peerStates,
		},
	}, nil
}

func (s *Service) GeneratePrivateAccount() (*ResultGeneratePrivateAccount, error) {
	privateAccount, err := acm.GeneratePrivateAccount()
	if err != nil {
		return nil, err
	}
	return &ResultGeneratePrivateAccount{
		PrivateAccount: privateAccount.ConcretePrivateAccount(),
	}, nil
}

func Status(blockchain bcm.BlockchainInfo, nodeView *tendermint.NodeView, blockTimeWithin, blockSeenTimeWithin string) (*ResultStatus, error) {
	publicKey := nodeView.ValidatorPublicKey()
	address := publicKey.Address()
	res := &ResultStatus{
		ChainID:       blockchain.ChainID(),
		RunID:         nodeView.RunID().String(),
		BurrowVersion: project.FullVersion(),
		GenesisHash:   blockchain.GenesisHash(),
		NodeInfo:      nodeView.NodeInfo(),
		SyncInfo: &SyncInfo{
			LatestBlockHeight:   blockchain.LastBlockHeight(),
			LatestBlockHash:     blockchain.LastBlockHash(),
			LatestAppHash:       blockchain.AppHashAfterLastBlock(),
			LatestBlockTime:     blockchain.LastBlockTime(),
			LatestBlockSeenTime: blockchain.LastCommitTime(),
			CatchingUp:          nodeView.IsFastSyncing(),
		},
		ValidatorInfo: &validator.Validator{
			Address:   &address,
			PublicKey: publicKey,
			Power:     blockchain.Validators().Power(address).Uint64(),
		},
	}

	now := time.Now()

	if blockTimeWithin != "" {
		err := timeWithin(now, res.SyncInfo.LatestBlockTime, blockTimeWithin)
		if err != nil {
			return nil, fmt.Errorf("have not committed block with sufficiently recent timestamp: %v, current status: %s",
				err, statusJSON(res))
		}
	}

	if blockSeenTimeWithin != "" {
		err := timeWithin(now, res.SyncInfo.LatestBlockSeenTime, blockSeenTimeWithin)
		if err != nil {
			return nil, fmt.Errorf("have not committed a block sufficiently recently: %v, current status: %s",
				err, statusJSON(res))
		}
	}

	return res, nil
}

func statusJSON(res *ResultStatus) string {
	bs, err := json.Marshal(res)
	if err != nil {
		bs = []byte("<error: could not marshal status>")
	}
	return string(bs)
}

func timeWithin(now time.Time, testTime time.Time, within string) error {
	duration, err := time.ParseDuration(within)
	if err != nil {
		return fmt.Errorf("could not parse duration '%s' to determine whether to throw error: %v", within, err)
	}
	// Take neg abs in case caller is counting backwards (note we later add the time since we normalise the duration to negative)
	if duration > 0 {
		duration = -duration
	}
	threshold := now.Add(duration)
	if testTime.After(threshold) {
		return nil
	}
	return fmt.Errorf("time %s does not fall within last %s (cutoff: %s)", testTime.Format(time.RFC3339), within,
		threshold.Format(time.RFC3339))
}
