package abci

import (
	"fmt"
	"time"

	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/project"
	"github.com/hyperledger/burrow/txs"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
)

const responseInfoName = "Burrow"

type abciApp struct {
	// State
	blockchain bcm.MutableBlockchain
	checker    execution.BatchExecutor
	committer  execution.BatchCommitter
	// We need to cache these from BeginBlock for when we need actually need it in Commit
	block *abci_types.RequestBeginBlock
	// Utility
	txDecoder txs.Decoder
	// Logging
	logger *logging.Logger
}

func NewApp(blockchain bcm.MutableBlockchain,
	checker execution.BatchExecutor,
	committer execution.BatchCommitter,
	logger *logging.Logger) abci_types.Application {
	return &abciApp{
		blockchain: blockchain,
		checker:    checker,
		committer:  committer,
		txDecoder:  txs.NewGoWireCodec(),
		logger:     logger.WithScope("abci.NewApp").With(structure.ComponentKey, "ABCI_App"),
	}
}

func (app *abciApp) Info(info abci_types.RequestInfo) abci_types.ResponseInfo {
	tip := app.blockchain.Tip()
	return abci_types.ResponseInfo{
		Data:             responseInfoName,
		Version:          project.History.CurrentVersion().String(),
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
	tx, err := app.txDecoder.DecodeTx(txBytes)
	if err != nil {
		app.logger.TraceMsg("CheckTx decoding error",
			"tag", "CheckTx",
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
		app.logger.TraceMsg("CheckTx execution error",
			structure.ErrorKey, err,
			"tag", "CheckTx",
			"tx_hash", receipt.TxHash,
			"creates_contract", receipt.CreatesContract)
		return abci_types.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("CheckTx could not execute transaction: %s, error: %v", tx, err),
		}
	}

	receiptBytes := wire.BinaryBytes(receipt)
	app.logger.TraceMsg("CheckTx success",
		"tag", "CheckTx",
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
	tx, err := app.txDecoder.DecodeTx(txBytes)
	if err != nil {
		app.logger.TraceMsg("DeliverTx decoding error",
			"tag", "DeliverTx",
			structure.ErrorKey, err)
		return abci_types.ResponseDeliverTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("Encoding error: %s", err),
		}
	}

	receipt := txs.GenerateReceipt(app.blockchain.ChainID(), tx)
	err = app.committer.Execute(tx)
	if err != nil {
		app.logger.TraceMsg("DeliverTx execution error",
			structure.ErrorKey, err,
			"tag", "DeliverTx",
			"tx_hash", receipt.TxHash,
			"creates_contract", receipt.CreatesContract)
		return abci_types.ResponseDeliverTx{
			Code: codes.TxExecutionErrorCode,
			Log:  fmt.Sprintf("DeliverTx could not execute transaction: %s, error: %s", tx, err),
		}
	}

	app.logger.TraceMsg("DeliverTx success",
		"tag", "DeliverTx",
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
	tip := app.blockchain.Tip()
	app.logger.InfoMsg("Committing block",
		"tag", "Commit",
		structure.ScopeKey, "Commit()",
		"block_height", app.block.Header.Height,
		"block_hash", app.block.Hash,
		"block_time", app.block.Header.Time,
		"num_txs", app.block.Header.NumTxs,
		"last_block_time", tip.LastBlockTime(),
		"last_block_hash", tip.LastBlockHash())

	// Commit state before resetting check cache so that the emptied cache servicing some RPC requests will fall through
	// to committed state when check state is reset
	appHash, err := app.committer.Commit()
	if err != nil {
		return abci_types.ResponseCommit{
			Code: codes.CommitErrorCode,
			Log:  fmt.Sprintf("Could not commit transactions in block to execution state: %s", err),
		}
	}

	// Commit to our blockchain state
	err = app.blockchain.CommitBlock(time.Unix(int64(app.block.Header.Time), 0), app.block.Hash, appHash)
	if err != nil {
		return abci_types.ResponseCommit{
			Code: codes.CommitErrorCode,
			Log:  fmt.Sprintf("Could not commit block to blockchain state: %s", err),
		}
	}

	err = app.checker.Reset()
	if err != nil {
		return abci_types.ResponseCommit{
			Code: codes.CommitErrorCode,
			Log:  fmt.Sprintf("Could not reset check cache during commit: %s", err),
		}
	}

	// Perform a sanity check our block height
	if app.blockchain.LastBlockHeight() != uint64(app.block.Header.Height) {
		app.logger.InfoMsg("Burrow block height disagrees with Tendermint block height",
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
