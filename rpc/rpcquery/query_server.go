package rpcquery

import (
	"context"
	"fmt"

	"github.com/hyperledger/burrow/txs/payload"

	"github.com/hyperledger/burrow/execution/proposal"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
)

type queryServer struct {
	accounts    acmstate.IterableStatsReader
	nameReg     names.IterableReader
	proposalReg proposal.IterableReader
	blockchain  bcm.BlockchainInfo
	nodeView    *tendermint.NodeView
	logger      *logging.Logger
}

var _ QueryServer = &queryServer{}

func NewQueryServer(state acmstate.IterableStatsReader, nameReg names.IterableReader, proposalReg proposal.IterableReader,
	blockchain bcm.BlockchainInfo, nodeView *tendermint.NodeView, logger *logging.Logger) *queryServer {
	return &queryServer{
		accounts:    state,
		nameReg:     nameReg,
		proposalReg: proposalReg,
		blockchain:  blockchain,
		nodeView:    nodeView,
		logger:      logger,
	}
}

func (qs *queryServer) Status(ctx context.Context, param *StatusParam) (*rpc.ResultStatus, error) {
	return rpc.Status(qs.blockchain, qs.nodeView, param.BlockTimeWithin, param.BlockSeenTimeWithin)
}

// Account state

func (qs *queryServer) GetAccount(ctx context.Context, param *GetAccountParam) (*acm.Account, error) {
	acc, err := qs.accounts.GetAccount(param.Address)
	if acc == nil {
		acc = &acm.Account{}
	}
	return acc, err
}

func (qs *queryServer) GetStorage(ctx context.Context, param *GetStorageParam) (*StorageValue, error) {
	val, err := qs.accounts.GetStorage(param.Address, param.Key)
	return &StorageValue{Value: val}, err
}

func (qs *queryServer) ListAccounts(param *ListAccountsParam, stream Query_ListAccountsServer) error {
	qry, err := query.NewOrEmpty(param.Query)
	var streamErr error
	err = qs.accounts.IterateAccounts(func(acc *acm.Account) error {
		if qry.Matches(acc.Tagged()) {
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

// Name registry
func (qs *queryServer) GetName(ctx context.Context, param *GetNameParam) (entry *names.Entry, err error) {
	entry, err = qs.nameReg.GetName(param.Name)
	if entry == nil && err == nil {
		err = fmt.Errorf("name %s not found", param.Name)
	}
	return
}

func (qs *queryServer) ListNames(param *ListNamesParam, stream Query_ListNamesServer) error {
	qry, err := query.NewOrEmpty(param.Query)
	if err != nil {
		return err
	}
	var streamErr error
	err = qs.nameReg.IterateNames(func(entry *names.Entry) error {
		if qry.Matches(entry.Tagged()) {
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

func (qs *queryServer) GetValidatorSet(ctx context.Context, param *GetValidatorSetParam) (*ValidatorSet, error) {
	set, deltas, height := qs.blockchain.ValidatorsHistory()
	vs := &ValidatorSet{
		Height: height,
		Set:    set.Validators(),
	}
	if param.IncludeHistory {
		vs.History = make([]*ValidatorSetDeltas, len(deltas))
		for i, d := range deltas {
			vs.History[i] = &ValidatorSetDeltas{
				Validators: d.Validators(),
			}
		}
	}
	return vs, nil
}

func (qs *queryServer) GetProposal(ctx context.Context, param *GetProposalParam) (proposal *payload.Ballot, err error) {
	proposal, err = qs.proposalReg.GetProposal(param.Hash)
	if proposal == nil && err == nil {
		err = fmt.Errorf("proposal %x not found", param.Hash)
	}
	return
}

func (qs *queryServer) ListProposals(param *ListProposalsParam, stream Query_ListProposalsServer) error {
	var streamErr error
	err := qs.proposalReg.IterateProposals(func(hash []byte, ballot *payload.Ballot) error {
		if param.GetProposed() == false || ballot.ProposalState == payload.Ballot_PROPOSED {
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
	stats := qs.accounts.GetAccountStats()

	return &Stats{
		AccountsWithCode:    stats.AccountsWithCode,
		AccountsWithoutCode: stats.AccountsWithoutCode,
	}, nil
}
