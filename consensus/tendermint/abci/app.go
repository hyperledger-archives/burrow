package abci

import (
	"fmt"
	"sync"
	"time"

	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/version"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
)

const responseInfoName = "Burrow"

type abciApp struct {
	mtx sync.Mutex
	// State
	blockchain bcm.MutableBlockchain
	checker    execution.BatchExecutor
	committer  execution.BatchCommitter
	// We need to cache these from BeginBlock for when we need actually need it in Commit
	block *abci_types.RequestBeginBlock
	// Utility
	txDecoder txs.Decoder
	// Logging
	logger logging_types.InfoTraceLogger
}

func NewApp(blockchain bcm.MutableBlockchain,
	checker execution.BatchExecutor,
	committer execution.BatchCommitter,
	logger logging_types.InfoTraceLogger) abci_types.Application {
	return &abciApp{
		blockchain: blockchain,
		checker:    checker,
		committer:  committer,
		txDecoder:  txs.NewGoWireCodec(),
		logger:     logging.WithScope(logger.With(structure.ComponentKey, "ABCI_App"), "abci.NewApp"),
	}
}

func (app *abciApp) Info(info abci_types.RequestInfo) abci_types.ResponseInfo {
	tip := app.blockchain.Tip()
	return abci_types.ResponseInfo{
		Data:             responseInfoName,
		Version:          version.GetSemanticVersionString(),
		LastBlockHeight:  int64(tip.LastBlockHeight()),
		LastBlockAppHash: tip.AppHashAfterLastBlock(),
	}
}

func (app *abciApp) SetOption(option abci_types.RequestSetOption) (respSetOption abci_types.ResponseSetOption) {
	respSetOption.Log = "SetOption not supported"
	respSetOption.Code = codes.UnsupportedRequestCode
	return
}

func (app *abciApp) Query(reqQuery abci_types.RequestQuery) (respQuery abci_types.ResponseQuery) {
	respQuery.Log = "Query not supported"
	respQuery.Code = codes.UnsupportedRequestCode
	return
}

func (app *abciApp) CheckTx(txBytes []byte) abci_types.ResponseCheckTx {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	tx, err := app.txDecoder.DecodeTx(txBytes)
	if err != nil {
		logging.TraceMsg(app.logger, "CheckTx decoding error",
			structure.ErrorKey, err)
		return abci_types.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("Encoding error: %s", err),
		}
	}
	// TODO: map ExecTx errors to sensible ABCI error codes
	receipt := txs.GenerateReceipt(app.blockchain.ChainID(), tx)

	err = app.checker.Execute(tx)
	if err != nil {
		logging.TraceMsg(app.logger, "CheckTx execution error",
			structure.ErrorKey, err,
			"tx_hash", receipt.TxHash,
			"creates_contract", receipt.CreatesContract)
		return abci_types.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("Could not execute transaction: %s, error: %v", tx, err),
		}
	}

	receiptBytes := wire.BinaryBytes(receipt)
	logging.TraceMsg(app.logger, "CheckTx success",
		"tx_hash", receipt.TxHash,
		"creates_contract", receipt.CreatesContract)
	return abci_types.ResponseCheckTx{
		Code: codes.TxExecutionSuccessCode,
		Log:  "CheckTx success - receipt in data",
		Data: receiptBytes,
	}
}

func (app *abciApp) InitChain(chain abci_types.RequestInitChain) (respInitChain abci_types.ResponseInitChain) {
	// Could verify agreement on initial validator set here
	return
}

func (app *abciApp) BeginBlock(block abci_types.RequestBeginBlock) (respBeginBlock abci_types.ResponseBeginBlock) {
	app.block = &block
	return
}

func (app *abciApp) DeliverTx(txBytes []byte) abci_types.ResponseDeliverTx {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	tx, err := app.txDecoder.DecodeTx(txBytes)
	if err != nil {
		logging.TraceMsg(app.logger, "DeliverTx decoding error",
			structure.ErrorKey, err)
		return abci_types.ResponseDeliverTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("Encoding error: %s", err),
		}
	}

	receipt := txs.GenerateReceipt(app.blockchain.ChainID(), tx)
	err = app.committer.Execute(tx)
	if err != nil {
		logging.TraceMsg(app.logger, "DeliverTx execution error",
			structure.ErrorKey, err,
			"tx_hash", receipt.TxHash,
			"creates_contract", receipt.CreatesContract)
		return abci_types.ResponseDeliverTx{
			Code: codes.TxExecutionErrorCode,
			Log:  fmt.Sprintf("Could not execute transaction: %s, error: %s", tx, err),
		}
	}

	logging.TraceMsg(app.logger, "DeliverTx success",
		"tx_hash", receipt.TxHash,
		"creates_contract", receipt.CreatesContract)
	receiptBytes := wire.BinaryBytes(receipt)
	return abci_types.ResponseDeliverTx{
		Code: codes.TxExecutionSuccessCode,
		Log:  "DeliverTx success - receipt in data",
		Data: receiptBytes,
	}
}

func (app *abciApp) EndBlock(reqEndBlock abci_types.RequestEndBlock) (respEndBlock abci_types.ResponseEndBlock) {
	// Validator mutation goes here
	return
}

func (app *abciApp) Commit() abci_types.ResponseCommit {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	tip := app.blockchain.Tip()
	logging.InfoMsg(app.logger, "Committing block",
		structure.ScopeKey, "Commit()",
		"block_height", tip.LastBlockHeight(),
		"block_hash", app.block.Hash,
		"block_time", app.block.Header.Time,
		"num_txs", app.block.Header.NumTxs,
		"last_block_time", tip.LastBlockTime(),
		"last_block_hash", tip.LastBlockHash())

	appHash, err := app.committer.Commit()
	if err != nil {
		return abci_types.ResponseCommit{
			Code: codes.CommitErrorCode,
			Log:  fmt.Sprintf("Could not commit block: %s", err),
		}
	}

	logging.InfoMsg(app.logger, "Resetting transaction check cache")
	app.checker.Reset()

	// Commit to our blockchain state
	app.blockchain.CommitBlock(time.Unix(int64(app.block.Header.Time), 0), app.block.Hash, appHash)

	// Perform a sanity check our block height
	if app.blockchain.LastBlockHeight() != uint64(app.block.Header.Height) {
		logging.InfoMsg(app.logger, "Burrow block height disagrees with Tendermint block height",
			structure.ScopeKey, "Commit()",
			"burrow_height", app.blockchain.LastBlockHeight(),
			"tendermint_height", app.block.Header.Height)
		return abci_types.ResponseCommit{
			Code: codes.CommitErrorCode,
			Log: fmt.Sprintf("Burrow has recorded a block height of %v, "+
				"but Tendermint reports a block height of %v, and the two should agree.",
				app.blockchain.LastBlockHeight(), app.block.Header.Height),
		}
	}
	return abci_types.ResponseCommit{
		Code: codes.TxExecutionSuccessCode,
		Data: appHash,
		Log:  "Success - AppHash in data",
	}
}
