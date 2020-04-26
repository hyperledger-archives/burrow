package lite

import (
	"context"

	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/tendermint/tendermint/lite2/provider"
	"github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
)

var _ provider.Provider = &Provider{}

type Provider struct {
	id  string
	cli rpcquery.QueryClient
}

func NewProvider(conn *grpc.ClientConn, chainID string) *Provider {
	return &Provider{
		id:  chainID,
		cli: rpcquery.NewQueryClient(conn),
	}
}

func (prov *Provider) ChainID() string {
	return prov.id
}

func (prov *Provider) SignedHeader(height int64) (*types.SignedHeader, error) {
	head, err := prov.cli.GetTendermintBlockHeader(context.Background(), &rpcquery.GetTendermintBlockHeaderParam{Height: height})
	if err != nil {
		return nil, err
	}
	return &types.SignedHeader{
		Header: &head.Header.Header,
		Commit: head.Commit.Commit,
	}, nil
}

func (prov *Provider) ValidatorSet(height int64) (*types.ValidatorSet, error) {
	set, err := prov.cli.GetTendermintValidatorSet(context.Background(), &rpcquery.GetTendermintValidatorSetParam{Height: height})
	if err != nil {
		return nil, err
	}

	return set.Set.ValidatorSet, nil
}
