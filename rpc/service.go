// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/registry"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/project"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/p2p"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

// Magic! Should probably be configurable, but not shouldn't be so huge we
// end up DoSing ourselves.
const MaxBlockLookback = 1000

// Base service that provides implementation for all underlying RPC methods
type Service struct {
	state      acmstate.IterableStatsReader
	nameReg    names.IterableReader
	nodeReg    registry.IterableReader
	blockchain bcm.BlockchainInfo
	validators validator.History
	nodeView   *tendermint.NodeView
	logger     *logging.Logger
}

// Service provides an internal query and information service with serialisable return types on which can accomodate
// a number of transport front ends
func NewService(state acmstate.IterableStatsReader, nameReg names.IterableReader, nodeReg registry.IterableReader, blockchain bcm.BlockchainInfo,
	validators validator.History, nodeView *tendermint.NodeView, logger *logging.Logger) *Service {

	return &Service{
		state:      state,
		nameReg:    nameReg,
		nodeReg:    nodeReg,
		blockchain: blockchain,
		validators: validators,
		nodeView:   nodeView,
		logger:     logger.With(structure.ComponentKey, "Service"),
	}
}

func (s *Service) Stats() acmstate.AccountStatsGetter {
	return s.state
}

func (s *Service) BlockchainInfo() bcm.BlockchainInfo {
	return s.blockchain
}

func (s *Service) ChainID() string {
	return s.blockchain.ChainID()
}

func (s *Service) UnconfirmedTxs(maxTxs int64) (*ResultUnconfirmedTxs, error) {
	if s.nodeView == nil {
		return nil, fmt.Errorf("cannot list unconfirmed transactions because NodeView not mounted")
	}
	// Get all transactions for now
	transactions, err := s.nodeView.MempoolTransactions(int(maxTxs))
	if err != nil {
		return nil, err
	}
	wrappedTxs := make([]*txs.Envelope, len(transactions))
	copy(wrappedTxs, transactions)
	return &ResultUnconfirmedTxs{
		NumTxs: len(transactions),
		Txs:    wrappedTxs,
	}, nil
}

func (s *Service) Status() (*ResultStatus, error) {
	return Status(s.BlockchainInfo(), s.validators, s.nodeView, "", "")
}

func (s *Service) StatusWithin(blockTimeWithin, blockSeenTimeWithin string) (*ResultStatus, error) {
	return Status(s.BlockchainInfo(), s.validators, s.nodeView, blockTimeWithin, blockSeenTimeWithin)
}

func (s *Service) ChainIdentifiers() (*ResultChainId, error) {
	return &ResultChainId{
		ChainName:   s.blockchain.GenesisDoc().ChainName,
		ChainId:     s.blockchain.ChainID(),
		GenesisHash: s.blockchain.GenesisHash(),
	}, nil
}

func (s *Service) Peers() []core_types.Peer {
	if s.nodeView == nil {
		return nil
	}
	p2pPeers := s.nodeView.Peers().List()
	peers := make([]core_types.Peer, len(p2pPeers))
	for i, peer := range p2pPeers {
		ni, _ := peer.NodeInfo().(p2p.DefaultNodeInfo)
		peers[i] = core_types.Peer{
			NodeInfo:         ni,
			IsOutbound:       peer.IsOutbound(),
			ConnectionStatus: peer.Status(),
		}
	}
	return peers
}

func (s *Service) Network() (*ResultNetwork, error) {
	if s.nodeView == nil {
		return nil, fmt.Errorf("cannot return network info because NodeView not mounted")
	}
	var listeners []string
	peers := s.Peers()
	return &ResultNetwork{
		ThisNode: s.nodeView.NodeInfo(),
		ResultNetInfo: &core_types.ResultNetInfo{
			Listening: true,
			Listeners: listeners,
			NPeers:    len(peers),
			Peers:     peers,
		},
	}, nil
}

func (s *Service) NetworkRegistry() ([]*ResultNetworkRegistry, error) {
	if s.nodeView == nil {
		return nil, fmt.Errorf("cannot return network registry info because NodeView not mounted")
	}

	rnr := make([]*ResultNetworkRegistry, 0)
	err := s.nodeReg.IterateNodes(func(id crypto.Address, rn *registry.NodeIdentity) error {
		rnr = append(rnr, &ResultNetworkRegistry{
			Address:      rn.ValidatorPublicKey.GetAddress(),
			NodeIdentity: *rn,
		})
		return nil
	})
	return rnr, err
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
	return &ResultAccount{Account: acc}, nil
}

func (s *Service) Accounts(predicate func(*acm.Account) bool) (*ResultAccounts, error) {
	accounts := make([]*acm.Account, 0)
	s.state.IterateAccounts(func(account *acm.Account) error {
		if predicate(account) {
			accounts = append(accounts, account)
		}
		return nil
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
	return &ResultStorage{Key: key, Value: value}, nil
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
	err = s.state.IterateStorage(address, func(key binary.Word256, value []byte) error {
		storageItems = append(storageItems, StorageItem{Key: key.UnpadLeft(), Value: value})
		return nil
	})
	if err != nil {
		return nil, err
	}
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
	tokens, err := acc.EVMCode.GetBytecode().Tokens()
	if err != nil {
		return nil, err
	}
	perms := permission.BasePermissionsToStringList(acc.Permissions.Base)

	return &ResultAccountHumanReadable{
		Account: &AccountHumanReadable{
			Address:     acc.GetAddress(),
			PublicKey:   acc.PublicKey,
			Sequence:    acc.Sequence,
			Balance:     acc.Balance,
			Code:        tokens,
			Permissions: perms,
			Roles:       acc.Permissions.Roles,
		},
	}, nil
}

func (s *Service) AccountStats() (*ResultAccountStats, error) {
	stats := s.state.GetAccountStats()
	return &ResultAccountStats{
		AccountsWithCode:    stats.AccountsWithCode,
		AccountsWithoutCode: stats.AccountsWithoutCode,
	}, nil
}

// Name registry
func (s *Service) Name(name string) (*ResultName, error) {
	entry, err := s.nameReg.GetName(name)
	if err != nil {
		return nil, err
	}
	return &ResultName{Entry: entry}, nil
}

func (s *Service) Names(predicate func(*names.Entry) bool) (*ResultNames, error) {
	var nms []*names.Entry
	err := s.nameReg.IterateNames(func(entry *names.Entry) error {
		if predicate(entry) {
			nms = append(nms, entry)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not iterate names: %v", err)
	}
	return &ResultNames{
		BlockHeight: s.blockchain.LastBlockHeight(),
		Names:       nms,
	}, nil
}

func (s *Service) Block(height uint64) (*ResultBlock, error) {
	if s.nodeView == nil {
		return nil, fmt.Errorf("NodeView is not mounted so cannot pull Tendermint blocks")
	}
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
	if s.nodeView == nil {
		return nil, fmt.Errorf("NodeView is not mounted so cannot pull Tendermint blocks")
	}
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
	for height := minHeight; height <= maxHeight; height++ {
		blockMeta := s.nodeView.BlockStore().LoadBlockMeta(height)
		blockMetas = append(blockMetas, blockMeta)
	}

	return &ResultBlocks{
		LastHeight: uint64(latestHeight),
		BlockMetas: blockMetas,
	}, nil
}

func (s *Service) Validators() (*ResultValidators, error) {
	var validators []*validator.Validator
	err := s.validators.Validators(0).IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		address := id.GetAddress()
		validators = append(validators, &validator.Validator{
			Address:   &address,
			PublicKey: id.GetPublicKey(),
			Power:     power.Uint64(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ResultValidators{
		BlockHeight:         s.blockchain.LastBlockHeight(),
		BondedValidators:    validators,
		UnbondingValidators: nil,
	}, nil
}

func (s *Service) ConsensusState() (*ResultConsensusState, error) {
	if s.nodeView == nil {
		return nil, fmt.Errorf("cannot pull ConsensusState because NodeView not mounted")
	}
	peers := s.nodeView.Peers().List()
	peerStates := make([]core_types.PeerStateInfo, len(peers))
	for i, peer := range peers {
		peerState := peer.Get(tmTypes.PeerStateKey).(*consensus.PeerState)
		peerStateJSON, err := peerState.ToJSON()
		if err != nil {
			return nil, err
		}
		netAddress, err := peer.NodeInfo().NetAddress()
		if err != nil {
			return nil, err
		}
		peerStates[i] = core_types.PeerStateInfo{
			// Peer basic info.
			NodeAddress: p2p.IDAddressString(peer.ID(), netAddress.String()),
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
	privateAccount, err := acm.GeneratePrivateAccount(crypto.CurveTypeEd25519)
	if err != nil {
		return nil, err
	}
	return &ResultGeneratePrivateAccount{
		PrivateAccount: privateAccount.ConcretePrivateAccount(),
	}, nil
}

func Status(blockchain bcm.BlockchainInfo, validators validator.History, nodeView *tendermint.NodeView, blockTimeWithin,
	blockSeenTimeWithin string) (*ResultStatus, error) {
	res := &ResultStatus{
		ChainID:       blockchain.ChainID(),
		RunID:         nodeView.RunID().String(),
		BurrowVersion: project.FullVersion(),
		GenesisHash:   blockchain.GenesisHash(),
		NodeInfo:      nodeView.NodeInfo(),
		SyncInfo:      bcm.GetSyncInfo(blockchain),
		CatchingUp:    nodeView.IsFastSyncing(),
	}
	if nodeView != nil {
		address := nodeView.ValidatorAddress()
		power, err := validators.Validators(0).Power(address)
		if err != nil {
			return nil, err
		}
		res.ValidatorInfo = &validator.Validator{
			Address:   &address,
			PublicKey: nodeView.ValidatorPublicKey(),
			Power:     power.Uint64(),
		}
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
