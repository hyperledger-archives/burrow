package proxy

import (
	"context"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/tendermint/tendermint/abci/types"
)

// Proxy Transact interface
func (p *Proxy) BroadcastTxSync(ctx context.Context, param *rpctransact.TxEnvelopeParam) (*exec.TxExecution, error) {
	return p.transact.BroadcastTxSync(ctx, param)
}

func (p *Proxy) BroadcastTxAsync(ctx context.Context, param *rpctransact.TxEnvelopeParam) (*txs.Receipt, error) {
	return p.transact.BroadcastTxAsync(ctx, param)
}

func (p *Proxy) SignTx(ctx context.Context, param *rpctransact.TxEnvelopeParam) (*rpctransact.TxEnvelope, error) {
	return p.transact.SignTx(ctx, param)
}

func (p *Proxy) FormulateTx(ctx context.Context, param *payload.Any) (*rpctransact.TxEnvelope, error) {
	return p.transact.FormulateTx(ctx, param)
}

func (p *Proxy) CallTxSync(ctx context.Context, param *payload.CallTx) (*exec.TxExecution, error) {
	return p.transact.CallTxSync(ctx, param)
}

func (p *Proxy) CallTxAsync(ctx context.Context, param *payload.CallTx) (*txs.Receipt, error) {
	return p.transact.CallTxAsync(ctx, param)
}

func (p *Proxy) CallTxSim(ctx context.Context, param *payload.CallTx) (*exec.TxExecution, error) {
	return p.transact.CallTxSim(ctx, param)
}

func (p *Proxy) CallCodeSim(ctx context.Context, param *rpctransact.CallCodeParam) (*exec.TxExecution, error) {
	return p.transact.CallCodeSim(ctx, param)
}

func (p *Proxy) SendTxSync(ctx context.Context, param *payload.SendTx) (*exec.TxExecution, error) {
	return p.transact.SendTxSync(ctx, param)
}

func (p *Proxy) SendTxAsync(ctx context.Context, param *payload.SendTx) (*txs.Receipt, error) {
	return p.transact.SendTxAsync(ctx, param)
}

func (p *Proxy) NameTxSync(ctx context.Context, param *payload.NameTx) (*exec.TxExecution, error) {
	return p.transact.NameTxSync(ctx, param)
}

func (p *Proxy) NameTxAsync(ctx context.Context, param *payload.NameTx) (*txs.Receipt, error) {
	return p.transact.NameTxAsync(ctx, param)
}

// Proxy Execution Events interface
func (p *Proxy) Tx(ctx context.Context, request *rpcevents.TxRequest) (*exec.TxExecution, error) {
	return p.events.Tx(ctx, request)
}

func (p *Proxy) Stream(request *rpcevents.BlocksRequest, stream rpcevents.ExecutionEvents_StreamServer) error {
	client, err := p.events.Stream(context.Background(), request)
	if err != nil {
		return err
	}

	for {
		acc, err := client.Recv()
		if err != nil {
			return err
		}
		err = stream.Send(acc)
		if err != nil {
			return err
		}
	}
}

func (p *Proxy) Events(request *rpcevents.BlocksRequest, stream rpcevents.ExecutionEvents_EventsServer) error {
	client, err := p.events.Events(context.Background(), request)
	if err != nil {
		return err
	}

	for {
		acc, err := client.Recv()
		if err != nil {
			return err
		}
		err = stream.Send(acc)
		if err != nil {
			return err
		}
	}
}

// Proxy Query interface
func (p *Proxy) Status(ctx context.Context, param *rpcquery.StatusParam) (*rpc.ResultStatus, error) {
	return p.query.Status(ctx, param)
}

func (p *Proxy) GetAccount(ctx context.Context, param *rpcquery.GetAccountParam) (*acm.Account, error) {
	return p.query.GetAccount(ctx, param)
}

func (p *Proxy) GetMetadata(ctx context.Context, param *rpcquery.GetMetadataParam) (*rpcquery.MetadataResult, error) {
	return p.query.GetMetadata(ctx, param)
}

func (p *Proxy) GetStorage(ctx context.Context, param *rpcquery.GetStorageParam) (*rpcquery.StorageValue, error) {
	return p.query.GetStorage(ctx, param)
}

func (p *Proxy) ListAccounts(param *rpcquery.ListAccountsParam, stream rpcquery.Query_ListAccountsServer) error {
	client, err := p.query.ListAccounts(context.Background(), param)
	if err != nil {
		return err
	}

	for {
		acc, err := client.Recv()
		if err != nil {
			return err
		}
		err = stream.Send(acc)
		if err != nil {
			return err
		}
	}
}

func (p *Proxy) GetName(ctx context.Context, param *rpcquery.GetNameParam) (entry *names.Entry, err error) {
	return p.query.GetName(ctx, param)
}

func (p *Proxy) ListNames(param *rpcquery.ListNamesParam, stream rpcquery.Query_ListNamesServer) error {
	client, err := p.query.ListNames(context.Background(), param)
	if err != nil {
		return err
	}

	for {
		acc, err := client.Recv()
		if err != nil {
			return err
		}
		err = stream.Send(acc)
		if err != nil {
			return err
		}
	}
}

func (p *Proxy) GetNetworkRegistry(ctx context.Context, param *rpcquery.GetNetworkRegistryParam) (*rpcquery.NetworkRegistry, error) {
	return p.query.GetNetworkRegistry(ctx, param)
}

func (p *Proxy) GetValidatorSet(ctx context.Context, param *rpcquery.GetValidatorSetParam) (*rpcquery.ValidatorSet, error) {
	return p.query.GetValidatorSet(ctx, param)
}

func (p *Proxy) GetValidatorSetHistory(ctx context.Context, param *rpcquery.GetValidatorSetHistoryParam) (*rpcquery.ValidatorSetHistory, error) {
	return p.query.GetValidatorSetHistory(ctx, param)
}

func (p *Proxy) GetProposal(ctx context.Context, param *rpcquery.GetProposalParam) (proposal *payload.Ballot, err error) {
	return p.query.GetProposal(ctx, param)
}

func (p *Proxy) ListProposals(param *rpcquery.ListProposalsParam, stream rpcquery.Query_ListProposalsServer) error {
	client, err := p.query.ListProposals(context.Background(), param)
	if err != nil {
		return err
	}

	for {
		acc, err := client.Recv()
		if err != nil {
			return err
		}
		err = stream.Send(acc)
		if err != nil {
			return err
		}
	}
}

func (p *Proxy) GetStats(ctx context.Context, param *rpcquery.GetStatsParam) (*rpcquery.Stats, error) {
	return p.query.GetStats(ctx, param)
}

func (p *Proxy) GetBlockHeader(ctx context.Context, param *rpcquery.GetBlockParam) (*types.Header, error) {
	return p.query.GetBlockHeader(ctx, param)
}

//
