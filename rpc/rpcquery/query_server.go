package rpcquery

import (
	"context"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/names"
)

type queryServer struct {
	accounts state.IterableReader
	nameReg  names.IterableReader
}

var _ QueryServer = &queryServer{}

func NewQueryServer(state state.IterableReader, nameReg names.IterableReader) *queryServer {
	return &queryServer{
		accounts: state,
		nameReg:  nameReg,
	}
}

func (qs *queryServer) GetAccount(ctx context.Context, param *GetAccountParam) (*acm.ConcreteAccount, error) {
	acc, err := qs.accounts.GetAccount(param.Address)
	if err != nil {
		return nil, err
	}
	return acm.AsConcreteAccount(acc), nil
}

func (qs *queryServer) ListAccounts(param *ListAccountsParam, stream Query_ListAccountsServer) error {
	qry, err := query.NewBuilder(param.Query).Query()
	var streamErr error
	_, err = qs.accounts.IterateAccounts(func(acc acm.Account) (stop bool) {
		if qry.Matches(acc.Tagged()) {
			streamErr = stream.Send(acm.AsConcreteAccount(acc))
			if streamErr != nil {
				return true
			}
		}
		return
	})
	if err != nil {
		return err
	}
	return streamErr
}

// Name registry
func (qs *queryServer) GetName(ctx context.Context, param *GetNameParam) (*names.Entry, error) {
	return qs.nameReg.GetName(param.Name)
}

func (qs *queryServer) ListNames(param *ListNamesParam, stream Query_ListNamesServer) error {
	qry, err := query.NewBuilder(param.Query).Query()
	if err != nil {
		return err
	}
	var streamErr error
	_, err = qs.nameReg.IterateNames(func(entry *names.Entry) (stop bool) {
		if qry.Matches(entry.Tagged()) {
			streamErr = stream.Send(entry)
			if streamErr != nil {
				return true
			}
		}
		return
	})
	if err != nil {
		return err
	}
	return streamErr
}
