package rpctransact

import (
	"fmt"
	"time"

	"github.com/hyperledger/burrow/logging"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"golang.org/x/net/context"
)

type Transactor interface {
	BroadcastTxSync(ctx context.Context, txEnv *txs.Envelope) (*exec.TxExecution, error)
	BroadcastTxAsync(ctx context.Context, txEnv *txs.Envelope) (*txs.Receipt, error)
	BroadcastTxStream(ctx context.Context, streamCtx context.Context, txEnv *txs.Envelope, consumer func(receipt *txs.Receipt, txe *exec.TxExecution) error) error
	ChainID() string
}

// This is probably silly
const maxBroadcastSyncTimeout = time.Hour

type transactServer struct {
	state      acmstate.Reader
	blockchain bcm.BlockchainInfo
	transactor Transactor
	txCodec    txs.Codec
	logger     *logging.Logger
}

func NewTransactServer(state acmstate.Reader, blockchain bcm.BlockchainInfo, transactor Transactor,
	txCodec txs.Codec, logger *logging.Logger) TransactServer {
	return &transactServer{
		state:      state,
		blockchain: blockchain,
		transactor: transactor,
		txCodec:    txCodec,
		logger:     logger.WithScope("NewTransactServer()"),
	}
}

func (ts *transactServer) BroadcastTxSync(ctx context.Context, param *TxEnvelopeParam) (*exec.TxExecution, error) {
	const errHeader = "BroadcastTxSync():"
	if param.Timeout == 0 {
		param.Timeout = maxBroadcastSyncTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, param.Timeout)
	defer cancel()
	txEnv := param.GetEnvelope(ts.transactor.ChainID())
	if txEnv == nil {
		return nil, fmt.Errorf("%s no transaction envelope or payload provided", errHeader)
	}
	return ts.transactor.BroadcastTxSync(ctx, txEnv)
}

func (ts *transactServer) BroadcastTxStream(param *TxEnvelopeParam, stream Transact_BroadcastTxStreamServer) error {
	const errHeader = "BroadcastTxStream():"
	if param.Timeout == 0 {
		param.Timeout = maxBroadcastSyncTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), param.Timeout)
	defer cancel()
	txEnv := param.GetEnvelope(ts.transactor.ChainID())
	if txEnv == nil {
		return fmt.Errorf("%s no transaction envelope or payload provided", errHeader)
	}
	return ts.transactor.BroadcastTxStream(ctx, stream.Context(), txEnv, func(receipt *txs.Receipt, txe *exec.TxExecution) error {
		return stream.Send(&BroadcastTxResult{Receipt: receipt, TxExecution: txe})
	})
}

func (ts *transactServer) BroadcastTxAsync(ctx context.Context, param *TxEnvelopeParam) (*txs.Receipt, error) {
	const errHeader = "BroadcastTxAsync():"
	if param.Timeout == 0 {
		param.Timeout = maxBroadcastSyncTimeout
	}
	txEnv := param.GetEnvelope(ts.transactor.ChainID())
	if txEnv == nil {
		return nil, fmt.Errorf("%s no transaction envelope or payload provided", errHeader)
	}
	return ts.transactor.BroadcastTxAsync(ctx, txEnv)
}

func (ts *transactServer) FormulateTx(ctx context.Context, param *payload.Any) (*TxEnvelope, error) {
	txEnv := txs.EnvelopeFromAny(ts.transactor.ChainID(), param)
	if txEnv == nil {
		return nil, fmt.Errorf("no payload provided to FormulateTx")
	}
	return &TxEnvelope{
		Envelope: txEnv,
	}, nil
}

func (ts *transactServer) CallTxSync(ctx context.Context, param *payload.CallTx) (*exec.TxExecution, error) {
	return ts.BroadcastTxSync(ctx, &TxEnvelopeParam{Payload: param.Any()})
}

func (ts *transactServer) CallTxAsync(ctx context.Context, param *payload.CallTx) (*txs.Receipt, error) {
	return ts.BroadcastTxAsync(ctx, &TxEnvelopeParam{Payload: param.Any()})
}

func (ts *transactServer) CallTxSim(ctx context.Context, param *payload.CallTx) (*exec.TxExecution, error) {
	if param.Address == nil {
		return nil, fmt.Errorf("CallSim requires a non-nil address from which to retrieve code")
	}
	return execution.CallSim(ts.state, ts.blockchain, param.Input.Address, *param.Address, param.Data, ts.logger)
}

func (ts *transactServer) CallCodeSim(ctx context.Context, param *CallCodeParam) (*exec.TxExecution, error) {
	return execution.CallCodeSim(ts.state, ts.blockchain, param.FromAddress, param.FromAddress, param.Code, param.Data,
		ts.logger)
}

func (ts *transactServer) SendTxSync(ctx context.Context, param *payload.SendTx) (*exec.TxExecution, error) {
	return ts.BroadcastTxSync(ctx, &TxEnvelopeParam{Payload: param.Any()})
}

func (ts *transactServer) SendTxAsync(ctx context.Context, param *payload.SendTx) (*txs.Receipt, error) {
	return ts.BroadcastTxAsync(ctx, &TxEnvelopeParam{Payload: param.Any()})
}

func (ts *transactServer) NameTxSync(ctx context.Context, param *payload.NameTx) (*exec.TxExecution, error) {
	return ts.BroadcastTxSync(ctx, &TxEnvelopeParam{Payload: param.Any()})
}

func (ts *transactServer) NameTxAsync(ctx context.Context, param *payload.NameTx) (*txs.Receipt, error) {
	return ts.BroadcastTxAsync(ctx, &TxEnvelopeParam{Payload: param.Any()})
}

func (te *TxEnvelopeParam) GetEnvelope(chainID string) *txs.Envelope {
	if te == nil {
		return nil
	}
	if te.Envelope != nil {
		return te.Envelope
	}
	if te.Payload != nil {
		return txs.EnvelopeFromAny(chainID, te.Payload)
	}
	return nil
}
