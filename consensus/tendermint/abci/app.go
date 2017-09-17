package abci

import (
	"sync"

	"fmt"

	"github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/version"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
	"time"
)

const responseInfoName = "Bosmarmot"

type abciApp struct {
	mtx sync.Mutex
	// State
	blockchain blockchain.Blockchain
	checker    execution.BatchExecutor
	committer  execution.BatchCommitter
	// We need to cache these from BeginBlock for when we need actually need it in Commit
	blockHash   []byte
	blockHeader *abci_types.Header
	// Utility
	txDecoder txs.Decoder
	logger    logging_types.InfoTraceLogger
}

func NewApp(blockchain blockchain.Blockchain,
	checker execution.BatchExecutor,
	committer execution.BatchCommitter,
	logger logging_types.InfoTraceLogger) abci_types.Application {
	return &abciApp{
		blockchain: blockchain,
		checker:    checker,
		committer:  committer,
		txDecoder:  txs.NewGoWireCodec(),
		logger:     logger,
	}
}

func (app *abciApp) Info() abci_types.ResponseInfo {
	return abci_types.ResponseInfo{
		Data:             responseInfoName,
		Version:          version.GetSemanticVersionString(),
		LastBlockHeight:  app.blockchain.LastBlockHeight(),
		LastBlockAppHash: app.blockchain.AppHashAfterLastBlock(),
	}
}

func (app *abciApp) SetOption(key string, value string) string {
	return "No options available"
}

func (app *abciApp) Query(reqQuery abci_types.RequestQuery) (respQuery abci_types.ResponseQuery) {
	respQuery.Log = "Query not support"
	respQuery.Code = abci_types.CodeType_UnknownRequest
	return respQuery
}

func (app *abciApp) CheckTx(txBytes []byte) abci_types.Result {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	tx, err := app.txDecoder.DecodeTx(txBytes)
	if err != nil {
		return abci_types.NewError(abci_types.CodeType_EncodingError, fmt.Sprintf("Encoding error: %v", err))
	}

	// TODO: map ExecTx errors to sensible ABCI error codes
	err = app.checker.Execute(tx)
	if err != nil {
		return abci_types.NewError(abci_types.CodeType_InternalError,
			fmt.Sprintf("Could not execute transaction: %s, error: %v", tx, err))
	}

	receiptBytes := wire.BinaryBytes(txs.GenerateReceipt(app.blockchain.ChainID(), tx))
	return abci_types.NewResultOK(receiptBytes, "Success")
}

func (app *abciApp) InitChain(validators []*abci_types.Validator) {
	// Could verify agreement on initial validator set here
}

func (app *abciApp) BeginBlock(hash []byte, header *abci_types.Header) {
	app.blockHeader = header
}

func (app *abciApp) DeliverTx(txBytes []byte) abci_types.Result {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	tx, err := app.txDecoder.DecodeTx(txBytes)
	if err != nil {
		return abci_types.NewError(abci_types.CodeType_EncodingError, fmt.Sprintf("Encoding error: %s", err))
	}

	err = app.committer.Execute(tx)
	if err != nil {
		return abci_types.NewError(abci_types.CodeType_InternalError,
			fmt.Sprintf("Could not execute transaction: %s, error: %s", tx, err))
	}

	receiptBytes := wire.BinaryBytes(txs.GenerateReceipt(app.blockchain.GenesisDoc().ChainID, tx))
	return abci_types.NewResultOK(receiptBytes, "Success")
}

func (app *abciApp) EndBlock(height uint64) (respEndBlock abci_types.ResponseEndBlock) {
	return respEndBlock
}

func (app *abciApp) Commit() abci_types.Result {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	logging.InfoMsg(app.logger, "Committing block",
		"last_block_height", app.blockchain.LastBlockHeight(),
		"last_block_time", app.blockchain.LastBlockTime(),
		"last_block_hash", app.blockchain.LastBlockHash())

	logging.InfoMsg(app.logger, "Resetting transaction check cache")
	app.checker.Reset()

	logging.InfoMsg(app.logger, "Committing transactions in block")
	appHash, err := app.committer.Commit()
	if err != nil {
		return abci_types.NewError(abci_types.CodeType_InternalError,
			fmt.Sprintf("Could not commit block: %s", err))
	}
	// Commit to our blockchain state
	app.blockchain.CommitBlock(time.Unix(int64(app.blockHeader.Time), 0), app.blockHash, appHash)
	return abci_types.NewResultOK(appHash, "Success")
}
