package abci

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"runtime/debug"

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
	checkTx       func(txBytes []byte) abciTypes.ResponseCheckTx
	deliverTx     func(txBytes []byte) abciTypes.ResponseCheckTx
	mempoolLocker sync.Locker
	// We need to cache these from BeginBlock for when we need actually need it in Commit
	block *abciTypes.RequestBeginBlock
	// Function to use to fail gracefully from panic rather than letting Tendermint make us a zombie
	panicFunc func(error)
	// Logging
	logger *logging.Logger
}

var _ abciTypes.Application = &App{}

func NewApp(blockchain *bcm.Blockchain, checker execution.BatchExecutor, committer execution.BatchCommitter,
	txDecoder txs.Decoder, panicFunc func(error), logger *logging.Logger) *App {
	return &App{
		blockchain: blockchain,
		checker:    checker,
		committer:  committer,
		checkTx:    txExecutor(checker, txDecoder, logger.WithScope("CheckTx")),
		deliverTx:  txExecutor(committer, txDecoder, logger.WithScope("DeliverTx")),
		panicFunc:  panicFunc,
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

func (app *App) InitChain(chain abciTypes.RequestInitChain) (respInitChain abciTypes.ResponseInitChain) {
	defer func() {
		if r := recover(); r != nil {
			app.panicFunc(fmt.Errorf("panic occurred in abci.App/InitChain: %v\n%s", r, debug.Stack()))
		}
	}()
	err := app.checkValidatorsMatch(chain.Validators)
	if err != nil {
		app.logger.InfoMsg("Initial validator set mistmatch", structure.ErrorKey, err)
		panic(err)
	}
	app.logger.InfoMsg("Initial validator set matches")
	return
}

func (app *App) BeginBlock(block abciTypes.RequestBeginBlock) (respBeginBlock abciTypes.ResponseBeginBlock) {
	app.block = &block
	defer func() {
		if r := recover(); r != nil {
			app.panicFunc(fmt.Errorf("panic occurred in abci.App/BeginBlock: %v\n%s", r, debug.Stack()))
		}
	}()
	if block.Header.Height > 1 {
		validators := make([]abciTypes.Validator, len(block.Validators))
		for i, v := range block.Validators {
			validators[i] = v.Validator
		}
		err := app.checkValidatorsMatch(validators)
		if err != nil {
			panic(err)
		}
	}
	return
}

func (app *App) checkValidatorsMatch(validators []abciTypes.Validator) error {
	tendermintValidators := bcm.NewValidators()
	for _, v := range validators {
		publicKey, err := crypto.PublicKeyFromABCIPubKey(v.PubKey)
		if err != nil {
			panic(err)
		}
		tendermintValidators.AlterPower(publicKey, big.NewInt(v.Power))
	}
	burrowValidators := app.blockchain.Validators()
	if !burrowValidators.Equal(tendermintValidators) {
		return fmt.Errorf("validators provided by Tendermint at InitChain do not match those held by Burrow: "+
			"Tendermint gives: %v, Burrow gives: %v", tendermintValidators, burrowValidators)
	}
	return nil
}

func (app *App) CheckTx(txBytes []byte) abciTypes.ResponseCheckTx {
	defer func() {
		if r := recover(); r != nil {
			app.panicFunc(fmt.Errorf("panic occurred in abci.App/CheckTx: %v\n%s", r, debug.Stack()))
		}
	}()
	return app.checkTx(txBytes)
}

func (app *App) DeliverTx(txBytes []byte) abciTypes.ResponseDeliverTx {
	defer func() {
		if r := recover(); r != nil {
			app.panicFunc(fmt.Errorf("panic occurred in abci.App/DeliverTx: %v\n%s", r, debug.Stack()))
		}
	}()
	ctr := app.deliverTx(txBytes)
	// Currently these message types are identical, if they are ever different can map between
	return abciTypes.ResponseDeliverTx{
		Code:      ctr.Code,
		Log:       ctr.Log,
		Data:      ctr.Data,
		Tags:      ctr.Tags,
		Fee:       ctr.Fee,
		GasUsed:   ctr.GasUsed,
		GasWanted: ctr.GasWanted,
		Info:      ctr.Info,
	}
}

func txExecutor(executor execution.BatchExecutor, txDecoder txs.Decoder, logger *logging.Logger) func(txBytes []byte) abciTypes.ResponseCheckTx {
	return func(txBytes []byte) abciTypes.ResponseCheckTx {
		txEnv, err := txDecoder.DecodeTx(txBytes)
		if err != nil {
			logger.TraceMsg("Decoding error",
				structure.ErrorKey, err)
			return abciTypes.ResponseCheckTx{
				Code: codes.EncodingErrorCode,
				Log:  fmt.Sprintf("Encoding error: %s", err),
			}
		}

		txe, err := executor.Execute(txEnv)
		if err != nil {
			logger.TraceMsg("Execution error",
				structure.ErrorKey, err,
				"tx_hash", txEnv.Tx.Hash())
			return abciTypes.ResponseCheckTx{
				Code: codes.EncodingErrorCode,
				Log:  fmt.Sprintf("Could not execute transaction: %s, error: %v", txEnv, err),
			}
		}

		bs, err := txe.Receipt.Encode()
		if err != nil {
			return abciTypes.ResponseCheckTx{
				Code: codes.TxExecutionErrorCode,
				Log:  fmt.Sprintf("Could not serialise receipt: %s", err),
			}
		}
		logger.TraceMsg("Execution success",
			"tx_hash", txe.TxHash,
			"contract_address", txe.Receipt.ContractAddress,
			"creates_contract", txe.Receipt.CreatesContract)
		return abciTypes.ResponseCheckTx{
			Code: codes.TxExecutionSuccessCode,
			Log:  "Execution success - TxExecution in data",
			Data: bs,
		}
	}
}

func (app *App) EndBlock(reqEndBlock abciTypes.RequestEndBlock) abciTypes.ResponseEndBlock {
	var validatorUpdates abciTypes.Validators
	app.blockchain.IterateValidators(func(id crypto.Addressable, power *big.Int) (stop bool) {
		validatorUpdates = append(validatorUpdates, abciTypes.Validator{
			Address: id.Address().Bytes(),
			PubKey:  id.PublicKey().ABCIPubKey(),
			// Must be ensured during execution
			Power: power.Int64(),
		})
		return
	})
	return abciTypes.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
	}
}

func (app *App) Commit() abciTypes.ResponseCommit {
	defer func() {
		if r := recover(); r != nil {
			app.panicFunc(fmt.Errorf("panic occurred in abci.App/Commit: %v\n%s", r, debug.Stack()))
		}
	}()
	blockTime := time.Unix(app.block.Header.Time, 0)
	app.logger.InfoMsg("Committing block",
		"tag", "Commit",
		structure.ScopeKey, "Commit()",
		"height", app.block.Header.Height,
		"hash", app.block.Hash,
		"txs", app.block.Header.NumTxs,
		"block_time", blockTime,
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

	appHash, err := app.committer.Commit(app.block.Hash, blockTime, &app.block.Header)
	if err != nil {
		panic(errors.Wrap(err, "Could not commit transactions in block to execution state"))
	}

	err = app.checker.Reset()
	if err != nil {
		panic(errors.Wrap(err, "could not reset check cache during commit"))
	}

	return abciTypes.ResponseCommit{
		Data: appHash,
	}
}
