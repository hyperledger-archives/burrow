package abci

import (
	"fmt"
	"sync"
	"time"

	"encoding/json"

	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/project"
	"github.com/hyperledger/burrow/txs"
	"github.com/pkg/errors"
	abciTypes "github.com/tendermint/abci/types"
)

const responseInfoName = "Burrow"

type App struct {
	// State
	blockchain    *bcm.Blockchain
	checker       execution.BatchExecutor
	committer     execution.BatchCommitter
	mempoolLocker sync.Locker
	// We need to cache these from BeginBlock for when we need actually need it in Commit
	block *abciTypes.RequestBeginBlock
	// Utility
	txDecoder txs.Decoder
	// Logging
	logger *logging.Logger
}

var _ abciTypes.Application = &App{}

func NewApp(blockchain *bcm.Blockchain,
	checker execution.BatchExecutor,
	committer execution.BatchCommitter,
	logger *logging.Logger) *App {
	return &App{
		blockchain: blockchain,
		checker:    checker,
		committer:  committer,
		txDecoder:  txs.NewJSONCodec(),
		logger:     logger.WithScope("abci.NewApp").With(structure.ComponentKey, "ABCI_App"),
	}
}

// Provide the Mempool lock. When provided we will attempt to acquire this lock in a goroutine during the Commit. We
// will keep the checker cache locked until we are able to acquire the mempool lock which signals the end of the commit
// and possible recheck on Tendermint's side.
func (app *App) SetMempoolLocker(mempoolLocker sync.Locker) {
	app.mempoolLocker = mempoolLocker
}

func (app *App) Info(info abciTypes.RequestInfo) abciTypes.ResponseInfo {
	tip := app.blockchain.Tip
	return abciTypes.ResponseInfo{
		Data:             responseInfoName,
		Version:          project.History.CurrentVersion().String(),
		LastBlockHeight:  int64(tip.LastBlockHeight()),
		LastBlockAppHash: tip.AppHashAfterLastBlock(),
	}
}

func (app *App) SetOption(option abciTypes.RequestSetOption) (respSetOption abciTypes.ResponseSetOption) {
	respSetOption.Log = "SetOption not supported"
	respSetOption.Code = codes.UnsupportedRequestCode
	return
}

func (app *App) Query(reqQuery abciTypes.RequestQuery) (respQuery abciTypes.ResponseQuery) {
	respQuery.Log = "Query not supported"
	respQuery.Code = codes.UnsupportedRequestCode
	return
}

func (app *App) CheckTx(txBytes []byte) abciTypes.ResponseCheckTx {
	txEnv, err := app.txDecoder.DecodeTx(txBytes)
	if err != nil {
		app.logger.TraceMsg("CheckTx decoding error",
			"tag", "CheckTx",
			structure.ErrorKey, err)
		return abciTypes.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("Encoding error: %s", err),
		}
	}
	receipt := txEnv.Tx.GenerateReceipt()

	err = app.checker.Execute(txEnv)
	if err != nil {
		app.logger.TraceMsg("CheckTx execution error",
			structure.ErrorKey, err,
			"tag", "CheckTx",
			"tx_hash", receipt.TxHash,
			"creates_contract", receipt.CreatesContract)
		return abciTypes.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("CheckTx could not execute transaction: %s, error: %v", txEnv, err),
		}
	}

	receiptBytes, err := json.Marshal(receipt)
	if err != nil {
		return abciTypes.ResponseCheckTx{
			Code: codes.TxExecutionErrorCode,
			Log:  fmt.Sprintf("CheckTx could not serialise receipt: %s", err),
		}
	}
	app.logger.TraceMsg("CheckTx success",
		"tag", "CheckTx",
		"tx_hash", receipt.TxHash,
		"creates_contract", receipt.CreatesContract)
	return abciTypes.ResponseCheckTx{
		Code: codes.TxExecutionSuccessCode,
		Log:  "CheckTx success - receipt in data",
		Data: receiptBytes,
	}
}

func (app *App) InitChain(chain abciTypes.RequestInitChain) (respInitChain abciTypes.ResponseInitChain) {
	// Could verify agreement on initial validator set here
	return
}

func (app *App) BeginBlock(block abciTypes.RequestBeginBlock) (respBeginBlock abciTypes.ResponseBeginBlock) {
	app.block = &block
	return
}

func (app *App) DeliverTx(txBytes []byte) abciTypes.ResponseDeliverTx {
	txEnv, err := app.txDecoder.DecodeTx(txBytes)
	if err != nil {
		app.logger.TraceMsg("DeliverTx decoding error",
			"tag", "DeliverTx",
			structure.ErrorKey, err)
		return abciTypes.ResponseDeliverTx{
			Code: codes.EncodingErrorCode,
			Log:  fmt.Sprintf("Encoding error: %s", err),
		}
	}

	receipt := txEnv.Tx.GenerateReceipt()
	err = app.committer.Execute(txEnv)
	if err != nil {
		app.logger.TraceMsg("DeliverTx execution error",
			structure.ErrorKey, err,
			"tag", "DeliverTx",
			"tx_hash", receipt.TxHash,
			"creates_contract", receipt.CreatesContract)
		return abciTypes.ResponseDeliverTx{
			Code: codes.TxExecutionErrorCode,
			Log:  fmt.Sprintf("DeliverTx could not execute transaction: %s, error: %s", txEnv, err),
		}
	}

	app.logger.TraceMsg("DeliverTx success",
		"tag", "DeliverTx",
		"tx_hash", receipt.TxHash,
		"creates_contract", receipt.CreatesContract)
	receiptBytes, err := json.Marshal(receipt)
	if err != nil {
		return abciTypes.ResponseDeliverTx{
			Code: codes.TxExecutionErrorCode,
			Log:  fmt.Sprintf("DeliverTx could not serialise receipt: %s", err),
		}
	}
	return abciTypes.ResponseDeliverTx{
		Code: codes.TxExecutionSuccessCode,
		Log:  "DeliverTx success - receipt in data",
		Data: receiptBytes,
	}
}

func (app *App) EndBlock(reqEndBlock abciTypes.RequestEndBlock) abciTypes.ResponseEndBlock {
	// Validator mutation goes here
	var validatorUpdates abciTypes.Validators
	app.blockchain.IterateValidators(func(publicKey crypto.PublicKey, power uint64) (stop bool) {
		validatorUpdates = append(validatorUpdates, abciTypes.Validator{
			Address: publicKey.Address().Bytes(),
			PubKey:  publicKey.ABCIPubKey(),
			Power:   int64(power),
		})
		return
	})
	return abciTypes.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
	}
}

func (app *App) Commit() abciTypes.ResponseCommit {
	app.logger.InfoMsg("Committing block",
		"tag", "Commit",
		structure.ScopeKey, "Commit()",
		"height", app.block.Header.Height,
		"hash", app.block.Hash,
		"txs", app.block.Header.NumTxs,
		"block_time", app.block.Header.Time, // [CSK] this sends a fairly non-sensical number; should be human readable
		"last_block_time", app.blockchain.Tip.LastBlockTime(),
		"last_block_hash", app.blockchain.Tip.LastBlockHash())

	// Lock the checker while we reset it and possibly while recheckTxs replays transactions
	app.checker.Lock()
	defer func() {
		// Tendermint may replay transactions to the check cache during a recheck, which happens after we have returned
		// from Commit(). The mempool is locked by Tendermint for the duration of the commit phase; during Commit() and
		// the subsequent mempool.Update() so we schedule an acquisition of the mempool lock in a goroutine in order to
		// 'observe' the mempool unlock event that happens later on. By keeping the checker read locked during that
		// period we can ensure that anything querying the checker (such as service.MempoolAccounts()) will block until
		// the full Tendermint commit phase has completed.
		if app.mempoolLocker != nil {
			go func() {
				// we won't get this until after the commit and we will acquire strictly after this commit phase has
				// ended (i.e. when Tendermint's BlockExecutor.Commit() returns
				app.mempoolLocker.Lock()
				// Prevent any mempool getting relocked while we unlock - we could just unlock immediately but if a new
				// commit starts gives goroutines blocked on checker a chance to progress before the next commit phase
				defer app.mempoolLocker.Unlock()
				app.checker.Unlock()
			}()
		} else {
			// If we have not be provided with access to the mempool lock
			app.checker.Unlock()
		}
	}()

	appHash, err := app.committer.Commit()
	if err != nil {
		panic(errors.Wrap(err, "Could not commit transactions in block to execution state"))

	}

	// Commit to our blockchain state
	err = app.blockchain.CommitBlock(time.Unix(int64(app.block.Header.Time), 0), app.block.Hash, appHash)
	if err != nil {
		panic(errors.Wrap(err, "could not commit block to blockchain state"))
	}

	err = app.checker.Reset()
	if err != nil {
		panic(errors.Wrap(err, "could not reset check cache during commit"))
	}

	// Perform a sanity check our block height
	if app.blockchain.LastBlockHeight() != uint64(app.block.Header.Height) {
		app.logger.InfoMsg("Burrow block height disagrees with Tendermint block height",
			structure.ScopeKey, "Commit()",
			"burrow_height", app.blockchain.LastBlockHeight(),
			"tendermint_height", app.block.Header.Height)

		panic(fmt.Errorf("burrow has recorded a block height of %v, "+
			"but Tendermint reports a block height of %v, and the two should agree",
			app.blockchain.LastBlockHeight(), app.block.Header.Height))
	}
	return abciTypes.ResponseCommit{
		Data: appHash,
	}
}
