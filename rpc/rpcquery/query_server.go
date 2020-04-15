package rpcquery

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/proposal"
	"github.com/hyperledger/burrow/execution/registry"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type queryServer struct {
	state      QueryState
	blockchain bcm.BlockchainInfo
	nodeView   *tendermint.NodeView
	logger     *logging.Logger
}

var _ QueryServer = &queryServer{}

type QueryState interface {
	acmstate.IterableStatsReader
	acmstate.MetadataReader
	names.IterableReader
	registry.IterableReader
	proposal.IterableReader
	validator.History
}

func NewQueryServer(state QueryState, blockchain bcm.BlockchainInfo, nodeView *tendermint.NodeView, logger *logging.Logger) *queryServer {
	return &queryServer{
		state:      state,
		blockchain: blockchain,
		nodeView:   nodeView,
		logger:     logger,
	}
}

func (qs *queryServer) Status(ctx context.Context, param *StatusParam) (*rpc.ResultStatus, error) {
	return rpc.Status(qs.blockchain, qs.state, qs.nodeView, param.BlockTimeWithin, param.BlockSeenTimeWithin)
}

// Account state

func (qs *queryServer) GetAccount(ctx context.Context, param *GetAccountParam) (*acm.Account, error) {
	acc, err := qs.state.GetAccount(param.Address)
	if acc == nil {
		acc = &acm.Account{}
	}
	return acc, err
}

// GetMetadata returns empty metadata string if not found. Metadata can be retrieved by account, or
// by metadata hash
func (qs *queryServer) GetMetadata(ctx context.Context, param *GetMetadataParam) (*MetadataResult, error) {
	metadata := &MetadataResult{}
	var contractMeta *acm.ContractMeta
	var err error
	if param.Address != nil {
		acc, err := qs.state.GetAccount(*param.Address)
		if err != nil {
			return metadata, err
		}
		if acc != nil && acc.CodeHash != nil {
			codehash := acc.CodeHash
			if acc.Forebear != nil {
				acc, err = qs.state.GetAccount(*acc.Forebear)
				if err != nil {
					return metadata, err
				}
			}

			for _, m := range acc.ContractMeta {
				if bytes.Equal(m.CodeHash, codehash) {
					contractMeta = m
					break
				}
			}

			if contractMeta == nil {
				deployCodehash := compile.GetDeployCodeHash(acc.EVMCode.GetBytecode(), *param.Address)
				for _, m := range acc.ContractMeta {
					if bytes.Equal(m.CodeHash, deployCodehash) {
						contractMeta = m
						break
					}
				}
			}
		}
	} else if param.MetadataHash != nil {
		contractMeta = &acm.ContractMeta{
			MetadataHash: *param.MetadataHash,
		}
	}
	if contractMeta == nil {
		return metadata, nil
	}
	if contractMeta.Metadata != "" {
		// Looks like the metadata is already memoised - (e.g. by native.State)
		metadata.Metadata = contractMeta.Metadata
	} else {
		var metadataHash acmstate.MetadataHash
		copy(metadataHash[:], contractMeta.MetadataHash)
		metadata.Metadata, err = qs.state.GetMetadata(metadataHash)
	}
	return metadata, err
}

func (qs *queryServer) GetStorage(ctx context.Context, param *GetStorageParam) (*StorageValue, error) {
	val, err := qs.state.GetStorage(param.Address, param.Key)
	return &StorageValue{Value: val}, err
}

func (qs *queryServer) ListAccounts(param *ListAccountsParam, stream Query_ListAccountsServer) error {
	qry, err := query.NewOrEmpty(param.Query)
	if err != nil {
		return err
	}
	var streamErr error
	err = qs.state.IterateAccounts(func(acc *acm.Account) error {
		if qry.Matches(acc) {
			return stream.Send(acc)
		} else {
			return nil
		}
	})
	if err != nil {
		return err
	}
	return streamErr
}

// Names

func (qs *queryServer) GetName(ctx context.Context, param *GetNameParam) (entry *names.Entry, err error) {
	entry, err = qs.state.GetName(param.Name)
	if entry == nil && err == nil {
		err = status.Error(codes.NotFound, fmt.Sprintf("name %s not found", param.Name))
	}
	return
}

func (qs *queryServer) ListNames(param *ListNamesParam, stream Query_ListNamesServer) error {
	qry, err := query.NewOrEmpty(param.Query)
	if err != nil {
		return err
	}
	var streamErr error
	err = qs.state.IterateNames(func(entry *names.Entry) error {
		if qry.Matches(entry) {
			return stream.Send(entry)
		} else {
			return nil
		}
	})
	if err != nil {
		return err
	}
	return streamErr
}

// Validators

func (qs *queryServer) GetValidatorSet(ctx context.Context, param *GetValidatorSetParam) (*ValidatorSet, error) {
	set := validator.Copy(qs.state.Validators(0))
	return &ValidatorSet{
		Set: set.Validators(),
	}, nil
}

func (qs *queryServer) GetValidatorSetHistory(ctx context.Context, param *GetValidatorSetHistoryParam) (*ValidatorSetHistory, error) {
	lookback := int(param.IncludePrevious)
	switch {
	case lookback == 0:
		lookback = 1
	case lookback < 0 || lookback > state.DefaultValidatorsWindowSize:
		lookback = state.DefaultValidatorsWindowSize
	}
	height := qs.blockchain.LastBlockHeight()
	if height < uint64(lookback) {
		lookback = int(height)
	}
	history := &ValidatorSetHistory{}
	for i := 0; i < lookback; i++ {
		set := validator.Copy(qs.state.Validators(i))
		vs := &ValidatorSet{
			Height: height - uint64(i),
			Set:    set.Validators(),
		}
		history.History = append(history.History, vs)
	}
	return history, nil
}

func (qs *queryServer) GetNetworkRegistry(ctx context.Context, param *GetNetworkRegistryParam) (*NetworkRegistry, error) {
	rv := make([]*RegisteredValidator, 0)
	err := qs.state.IterateNodes(func(id crypto.Address, rn *registry.NodeIdentity) error {
		rv = append(rv, &RegisteredValidator{
			Address: rn.ValidatorPublicKey.GetAddress(),
			Node:    rn,
		})
		return nil
	})
	return &NetworkRegistry{Set: rv}, err
}

// Proposals

func (qs *queryServer) GetProposal(ctx context.Context, param *GetProposalParam) (proposal *payload.Ballot, err error) {
	proposal, err = qs.state.GetProposal(param.Hash)
	if proposal == nil && err == nil {
		err = fmt.Errorf("proposal %x not found", param.Hash)
	}
	return
}

func (qs *queryServer) ListProposals(param *ListProposalsParam, stream Query_ListProposalsServer) error {
	var streamErr error
	err := qs.state.IterateProposals(func(hash []byte, ballot *payload.Ballot) error {
		if !param.GetProposed() || ballot.ProposalState == payload.Ballot_PROPOSED {
			return stream.Send(&ProposalResult{Hash: hash, Ballot: ballot})
		} else {
			return nil
		}
	})
	if err != nil {
		return err
	}
	return streamErr
}

func (qs *queryServer) GetStats(ctx context.Context, param *GetStatsParam) (*Stats, error) {
	stats := qs.state.GetAccountStats()

	return &Stats{
		AccountsWithCode:    stats.AccountsWithCode,
		AccountsWithoutCode: stats.AccountsWithoutCode,
	}, nil
}

// Tendermint and blocks

func (qs *queryServer) GetBlockHeader(ctx context.Context, param *GetBlockParam) (*types.Header, error) {
	header, err := qs.blockchain.GetBlockHeader(param.Height)
	if err != nil {
		return nil, err
	}
	abciHeader := tmtypes.TM2PB.Header(header)
	return &abciHeader, nil
}
