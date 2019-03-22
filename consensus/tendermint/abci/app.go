package abci

import (
	"fmt"
	"math/big"
	"runtime/debug"
	"sync"

	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/project"
	"github.com/hyperledger/burrow/txs"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

type Validators interface {
	validator.History
}

const (
	TendermintValidatorDelayInBlocks = 2
	BurrowValidatorDelayInBlocks     = 1
)

type App struct {
	// Node information to return in Info
	nodeInfo string
	// State
	blockchain              *bcm.Blockchain
	validators              Validators
	checkTx                 func(txBytes []byte) abciTypes.ResponseCheckTx
	deliverTx               func(txBytes []byte) abciTypes.ResponseCheckTx
	mempoolLocker           sync.Locker
	authorizedPeersProvider PeersFilterProvider
	// We need to cache these from BeginBlock for when we need actually need it in Commit
	block *abciTypes.RequestBeginBlock
	// Function to use to fail gracefully from panic rather than letting Tendermint make us a zombie
	panicFunc func(error)
	TxExecutor
}

type TxExecutor struct {
	checker   execution.BatchExecutor
	committer execution.BatchCommitter
	txDecoder txs.Decoder
	logger    *logging.Logger
}

// PeersFilterProvider provides current authorized nodes id and/or addresses
type PeersFilterProvider func() (authorizedPeersID []string, authorizedPeersAddress []string)

var _ abciTypes.Application = &App{}

func NewApp(nodeInfo string, blockchain *bcm.Blockchain, validators Validators, checker execution.BatchExecutor,
	committer execution.BatchCommitter, txDecoder txs.Decoder, authorizedPeersProvider PeersFilterProvider,
	panicFunc func(error), logger *logging.Logger) *App {
	return &App{
		nodeInfo:   nodeInfo,
		blockchain: blockchain,
		validators: validators,
		TxExecutor: NewTxExecutor(nodeInfo, checker, committer, txDecoder,
			logger.WithScope("abci.NewApp").With(structure.ComponentKey, "ABCI_App", "node_info", nodeInfo)),
		authorizedPeersProvider: authorizedPeersProvider,
		panicFunc:               panicFunc,
	}
}

func NewTxExecutor(nodeInfo string, checker execution.BatchExecutor, committer execution.BatchCommitter,
	txDecoder txs.Decoder, logger *logging.Logger) TxExecutor {
	return TxExecutor{
		checker,
		committer,
		txDecoder,
		logger.WithScope("abci.NewApp").With(structure.ComponentKey, "ABCI_App", "node_info", nodeInfo),
	}
}

// Provide the Mempool lock. When provided we will attempt to acquire this lock in a goroutine during the Commit. We
// will keep the checker cache locked until we are able to acquire the mempool lock which signals the end of the commit
// and possible recheck on Tendermint's side.
func (app *App) SetMempoolLocker(mempoolLocker sync.Locker) {
	app.mempoolLocker = mempoolLocker
}

func (app *App) Info(info abciTypes.RequestInfo) abciTypes.ResponseInfo {
	return abciTypes.ResponseInfo{
		Data:             app.nodeInfo,
		Version:          project.History.CurrentVersion().String(),
		LastBlockHeight:  int64(app.blockchain.LastBlockHeight()),
		LastBlockAppHash: app.blockchain.AppHashAfterLastBlock(),
	}
}

func (app *App) SetOption(option abciTypes.RequestSetOption) (respSetOption abciTypes.ResponseSetOption) {
	respSetOption.Log = "SetOption not supported"
	respSetOption.Code = codes.UnsupportedRequestCode
	return
}

func (app *App) Query(reqQuery abciTypes.RequestQuery) (respQuery abciTypes.ResponseQuery) {
	defer func() {
		if r := recover(); r != nil {
			app.panicFunc(fmt.Errorf("panic occurred in abci.App/Query: %v\n%s", r, debug.Stack()))
		}
	}()
	respQuery.Log = "Query not supported"
	respQuery.Code = codes.UnsupportedRequestCode

	switch {
	case isPeersFilterQuery(&reqQuery):
		app.peersFilter(&reqQuery, &respQuery)
	}
	return
}

func (app *App) InitChain(chain abciTypes.RequestInitChain) (respInitChain abciTypes.ResponseInitChain) {
	defer func() {
		if r := recover(); r != nil {
			app.panicFunc(fmt.Errorf("panic occurred in abci.App/InitChain: %v\n%s", r, debug.Stack()))
		}
	}()
	currentSet := validator.NewTrimSet()
	err := validator.Write(currentSet, app.validators.Validators(0))
	if err != nil {
		panic(fmt.Errorf("could not build current validator set: %v", err))
	}
	if len(chain.Validators) != currentSet.Size() {
		panic(fmt.Errorf("Tendermint passes %d validators to InitChain but Burrow's Blockchain has %d",
			len(chain.Validators), currentSet.Size()))
	}
	for _, v := range chain.Validators {
		pk, err := crypto.PublicKeyFromABCIPubKey(v.GetPubKey())
		err = app.checkValidatorMatches(currentSet, abciTypes.Validator{Address: pk.GetAddress().Bytes(), Power: v.Power})
		if err != nil {
			panic(err)
		}
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
		var err error
		previousValidators := validator.NewTrimSet()
		// Tendermint runs two blocks behind plus we are updating in end block validators updated last round
		err = validator.Write(previousValidators,
			app.validators.Validators(BurrowValidatorDelayInBlocks+TendermintValidatorDelayInBlocks))
		if err != nil {
			panic(fmt.Errorf("could not build current validator set: %v", err))
		}
		if len(block.LastCommitInfo.Votes) != previousValidators.Size() {
			err = fmt.Errorf("Tendermint passes %d validators to BeginBlock but Burrow's has %d:\n %v",
				len(block.LastCommitInfo.Votes), previousValidators.Size(), previousValidators.String())
			panic(err)
		}
		for _, v := range block.LastCommitInfo.Votes {
			err = app.checkValidatorMatches(previousValidators, v.Validator)
			if err != nil {
				panic(err)
			}
		}
	}
	return
}

func (app *App) checkValidatorMatches(ours validator.Reader, v abciTypes.Validator) error {
	address, err := crypto.AddressFromBytes(v.Address)
	if err != nil {
		return err
	}
	power, err := ours.Power(address)
	if err != nil {
		return err
	}
	if power.Cmp(big.NewInt(v.Power)) != 0 {
		return fmt.Errorf("validator %v has power %d from Tendermint but power %d from Burrow",
			address, v.Power, power)
	}
	return nil
}

func (app *App) CheckTx(txBytes []byte) abciTypes.ResponseCheckTx {
	defer func() {
		if r := recover(); r != nil {
			app.panicFunc(fmt.Errorf("panic occurred in abci.App/CheckTx: %v\n%s", r, debug.Stack()))
		}
	}()
	return app.TxExecutor.CheckTx(txBytes)
}

func (app *App) DeliverTx(txBytes []byte) abciTypes.ResponseDeliverTx {
	defer func() {
		if r := recover(); r != nil {
			app.panicFunc(fmt.Errorf("panic occurred in abci.App/DeliverTx: %v\n%s", r, debug.Stack()))
		}
	}()
	return app.TxExecutor.DeliverTx(txBytes)
}

func (txx TxExecutor) CheckTx(txBytes []byte) abciTypes.ResponseCheckTx {
	return txx.execute("CheckTx", txx.checker, txBytes)
}

func (txx TxExecutor) DeliverTx(txBytes []byte) abciTypes.ResponseDeliverTx {
	ctr := txx.execute("DeliverTx", txx.committer, txBytes)
	// Currently these message types are identical, if they are ever different can map between
	return abciTypes.ResponseDeliverTx{
		Code:      ctr.Code,
		Log:       ctr.Log,
		Data:      ctr.Data,
		Tags:      ctr.Tags,
		GasUsed:   ctr.GasUsed,
		GasWanted: ctr.GasWanted,
		Info:      ctr.Info,
	}
}

func (txx TxExecutor) execute(name string, executor execution.BatchExecutor, txBytes []byte) abciTypes.ResponseCheckTx {
	logf := func(format string, args ...interface{}) string {
		return fmt.Sprintf("%s: "+format, append([]interface{}{name}, args...)...)
	}

	txEnv, err := txx.txDecoder.DecodeTx(txBytes)
	if err != nil {
		txx.logger.InfoMsg("Decoding error",
			structure.ErrorKey, err)
		return abciTypes.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Log:  logf("Encoding error: %s", err),
		}
	}

	txe, err := executor.Execute(txEnv)
	if err != nil {
		ex := errors.AsException(err)
		txx.logger.InfoMsg("Execution error",
			structure.ErrorKey, err,
			"tx_hash", txEnv.Tx.Hash())
		return abciTypes.ResponseCheckTx{
			Code: codes.TxExecutionErrorCode,
			Log:  logf("Could not execute transaction: %s, error: %v", txEnv, ex.Exception),
		}
	}

	bs, err := txe.Receipt.Encode()
	if err != nil {
		return abciTypes.ResponseCheckTx{
			Code: codes.EncodingErrorCode,
			Log:  logf("Could not serialise receipt: %s", err),
		}
	}
	txx.logger.InfoMsg("Execution success",
		"tx_hash", txe.TxHash,
		"contract_address", txe.Receipt.ContractAddress,
		"creates_contract", txe.Receipt.CreatesContract)
	return abciTypes.ResponseCheckTx{
		Code: codes.TxExecutionSuccessCode,
		Log:  logf("Execution success - TxExecution in data"),
		Data: bs,
	}

}

func (app *App) EndBlock(reqEndBlock abciTypes.RequestEndBlock) abciTypes.ResponseEndBlock {
	var validatorUpdates []abciTypes.ValidatorUpdate
	defer func() {
		if r := recover(); r != nil {
			app.panicFunc(fmt.Errorf("panic occurred in abci.App/EndBlock: %v\n%s", r, debug.Stack()))
		}
	}()
	err := app.validators.ValidatorChanges(BurrowValidatorDelayInBlocks).IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		app.logger.InfoMsg("Updating validator power", "validator_address", id.GetAddress(),
			"new_power", power)
		validatorUpdates = append(validatorUpdates, abciTypes.ValidatorUpdate{
			PubKey: id.GetPublicKey().ABCIPubKey(),
			// Must ensure power fits in an int64 during execution
			Power: power.Int64(),
		})
		return nil
	})
	if err != nil {
		panic(err)
	}
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
	blockTime := app.block.Header.Time
	app.logger.InfoMsg("Committing block",
		"tag", "Commit",
		structure.ScopeKey, "Commit()",
		"height", app.block.Header.Height,
		"hash", app.block.Hash,
		"txs", app.block.Header.NumTxs,
		"block_time", blockTime,
		"last_block_time", app.blockchain.LastBlockTime(),
		"last_block_duration", app.blockchain.LastCommitDuration(),
		"last_block_hash", app.blockchain.LastBlockHash(),
	)

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

	appHash, err := app.committer.Commit(&app.block.Header)
	if err != nil {
		panic(errors.Wrap(err, "Could not commit transactions in block to execution state"))
	}
	err = app.checker.Reset()
	if err != nil {
		panic(errors.Wrap(err, "could not reset check cache during commit"))
	}
	// Commit to our blockchain state which will checkpoint the previous app hash by saving it to the database
	// (we know the previous app hash is safely committed because we are about to commit the next)
	err = app.blockchain.CommitBlock(blockTime, app.block.Hash, appHash)
	if err != nil {
		panic(fmt.Errorf("could not commit block to blockchain state: %v", err))
	}
	app.logger.InfoMsg("Committed block")

	return abciTypes.ResponseCommit{
		Data: appHash,
	}
}
