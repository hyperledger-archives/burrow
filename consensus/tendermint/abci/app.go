package abci

import (
	"sync"

	"fmt"

	"time"

	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/version"
	"github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/events"
	"github.com/hyperledger/burrow/genesis"
)

const responseInfoName = "Bosmarmot"

type abciApp struct {
	// State commit mutex,
	mtx       sync.Mutex
	AppHash   []byte
	BlockTime time.Time

	state      *execution.State
	cache      *execution.BlockCache
	checkCache *execution.BlockCache // for CheckTx (eg. so we get nonces right)

	// TODO: get rid of this, don't buffer here that's not down to us.
	eventCache  *events.EventCache
	eventSwitch events.EventSwitch

	logger logging_types.InfoTraceLogger
}

func NewApp(genesisDoc genesis.GenesisDoc, state *execution.State, eventSwitch events.EventSwitch,
	logger logging_types.InfoTraceLogger) types.Application {
	return &abciApp{
		AppHash: genesisDoc.AppHash,
		BlockTime: genesisDoc.GenesisTime,
		state: state,
		eventSwitch: eventSwitch,
	}
}

func (app *abciApp) Info() types.ResponseInfo {
	return types.ResponseInfo{
		Data:             responseInfoName,
		Version:          version.GetSemanticVersionString(),
		LastBlockHeight:  uint64(app.state.LastBlockHeight),
		LastBlockAppHash: app.state.LastBlockAppHash,
	}
}

func (app *abciApp) SetOption(key string, value string) string {
	return "No options available"
}

func (app *abciApp) Query(reqQuery types.RequestQuery) (respQuery types.ResponseQuery) {
	respQuery.Log = "Query not support"
	respQuery.Code = types.CodeType_UnknownRequest
	return respQuery
}

func (app *abciApp) CheckTx(txBytes []byte) types.Result {
	tx, err := txs.DecodeTx(txBytes)
	if err != nil {
		return types.NewError(types.CodeType_EncodingError, fmt.Sprintf("Encoding error: %v", err))
	}

	// TODO: map ExecTx errors to sensible ABCI error codes
	err = state.ExecTx(app.checkCache, tx, false, nil, app.logger)
	if err != nil {
		return types.NewError(types.CodeType_InternalError, fmt.Sprintf("Internal error: %v", err))
	}
	receipt := txs.GenerateReceipt(app.state.ChainID, tx)
	receiptBytes := wire.BinaryBytes(receipt)
	return types.NewResultOK(receiptBytes, "Success")
}

func (app *abciApp) InitChain(validators []*types.Validator) {
	// Could verify agreement on initial validator set here
}

func (app *abciApp) BeginBlock(hash []byte, header *types.Header) {
	app.BlockTime = time.Unix(int64(header.Time), 0)
}

func (app *abciApp) DeliverTx(txBytes []byte) types.Result {
	tx, err := txs.DecodeTx(txBytes)
	if err != nil {
		return types.NewError(types.CodeType_EncodingError, fmt.Sprintf("Encoding error: %v", err))
	}

	err = state.ExecTx(app.cache, tx, true, app.eventCache, app.logger)
	if err != nil {
		return types.NewError(types.CodeType_InternalError, fmt.Sprintf("Internal error: %v", err))
	}

	receipt := txs.GenerateReceipt(app.state.ChainID, tx)
	receiptBytes := wire.BinaryBytes(receipt)
	return types.NewResultOK(receiptBytes, "Success")
}

func (app *abciApp) EndBlock(height uint64) (respEndBlock types.ResponseEndBlock) {
	return respEndBlock
}

func (app *abciApp) Commit() types.Result {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	logging.InfoMsg(app.logger, "Committing block",
		"last_block_height", app.state.LastBlockHeight,
		"last_block_time", app.state.LastBlockTime,
		"last_block_app_hash", app.state.LastBlockAppHash)

	// sync the AppendTx cache
	app.cache.Sync()

	// Refresh the checkCache with the latest commited state
	logging.InfoMsg(app.logger, "Resetting checkCache")
	app.checkCache = execution.NewBlockCache(app.state)

	// flush events to listeners (XXX: note issue with blocking)
	app.eventCache.Flush()

	// The versions of these stored in app were updated in BeginBLock
	app.state.LastBlockHeight += 1
	app.state.LastBlockTime = app.BlockTime
	app.state.LastBlockAppHash = app.AppHash
	// save state to disk
	app.state.Save()

	app.AppHash = app.state.Hash()
	return types.NewResultOK(app.AppHash, "Success")
}
