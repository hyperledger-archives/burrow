// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package execution

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/contexts"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/proposal"
	"github.com/hyperledger/burrow/execution/registry"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/tendermint/tendermint/proto/tendermint/types"
)

type Executor interface {
	Execute(txEnv *txs.Envelope) (*exec.TxExecution, error)
}

type ExecutorFunc func(txEnv *txs.Envelope) (*exec.TxExecution, error)

func (f ExecutorFunc) Execute(txEnv *txs.Envelope) (*exec.TxExecution, error) {
	return f(txEnv)
}

type ExecutorState interface {
	Update(updater func(ws state.Updatable) error) (hash []byte, version int64, err error)
	LastStoredHeight() (uint64, error)
	acmstate.IterableReader
	acmstate.MetadataReader
	names.Reader
	registry.Reader
	proposal.Reader
	validator.IterableReader
}
type BatchExecutor interface {
	// Provides access to write lock for a BatchExecutor so reads can be prevented for the duration of a commit
	sync.Locker
	// Used by execution.Accounts to implement memory pool signing
	acmstate.Reader
	// Execute transaction against block cache (i.e. block buffer)
	Executor
	// Reset executor to underlying State
	Reset() error
}

// Executes transactions
type BatchCommitter interface {
	BatchExecutor
	// Commit execution results to underlying State and provide opportunity to mutate state before it is saved
	Commit(header *types.Header) (stateHash []byte, err error)
}

type executor struct {
	sync.RWMutex
	runCall          bool
	params           Params
	state            ExecutorState
	stateCache       *acmstate.Cache
	metadataCache    *acmstate.MetadataCache
	nameRegCache     *names.Cache
	nodeRegCache     *registry.Cache
	proposalRegCache *proposal.Cache
	validatorCache   *validator.Cache
	emitter          *event.Emitter
	block            *exec.BlockExecution
	logger           *logging.Logger
	vmOptions        evm.Options
	contexts         map[payload.Type]contexts.Context
}

type Params struct {
	ChainID           string
	ProposalThreshold uint64
}

func ParamsFromGenesis(genesisDoc *genesis.GenesisDoc) Params {
	return Params{
		ChainID:           genesisDoc.ChainID(),
		ProposalThreshold: genesisDoc.Params.ProposalThreshold,
	}
}

var _ BatchExecutor = (*executor)(nil)

// Wraps a cache of what is variously known as the 'check cache' and 'mempool'
func NewBatchChecker(backend ExecutorState, params Params, blockchain engine.Blockchain, logger *logging.Logger,
	options ...Option) (BatchExecutor, error) {

	return newExecutor("CheckCache", false, params, backend, blockchain, nil,
		logger.WithScope("NewBatchExecutor"), options...)
}

func NewBatchCommitter(backend ExecutorState, params Params, blockchain engine.Blockchain, emitter *event.Emitter,
	logger *logging.Logger, options ...Option) (BatchCommitter, error) {

	return newExecutor("CommitCache", true, params, backend, blockchain, emitter,
		logger.WithScope("NewBatchCommitter"), options...)

}

func newExecutor(name string, runCall bool, params Params, backend ExecutorState, blockchain engine.Blockchain,
	emitter *event.Emitter, logger *logging.Logger, options ...Option) (*executor, error) {
	// We need to track the last block stored in state
	predecessor, err := backend.LastStoredHeight()
	if err != nil {
		return nil, err
	}
	exe := &executor{
		runCall:          runCall,
		params:           params,
		state:            backend,
		stateCache:       acmstate.NewCache(backend, acmstate.Named(name)),
		metadataCache:    acmstate.NewMetadataCache(backend),
		nameRegCache:     names.NewCache(backend),
		nodeRegCache:     registry.NewCache(backend),
		proposalRegCache: proposal.NewCache(backend),
		validatorCache:   validator.NewCache(backend),
		emitter:          emitter,
		block: &exec.BlockExecution{
			Height:            blockchain.LastBlockHeight() + 1,
			PredecessorHeight: predecessor,
		},
		logger: logger.With(structure.ComponentKey, "Executor"),
	}
	for _, option := range options {
		option(exe)
	}

	baseContexts := map[payload.Type]contexts.Context{
		payload.TypeCall: &contexts.CallContext{
			EVM:           evm.New(exe.vmOptions),
			Blockchain:    blockchain,
			State:         exe.stateCache,
			MetadataState: exe.metadataCache,
			RunCall:       runCall,
			Logger:        exe.logger,
		},
		payload.TypeSend: &contexts.SendContext{
			State:  exe.stateCache,
			Logger: exe.logger,
		},
		payload.TypeName: &contexts.NameContext{
			Blockchain: blockchain,
			State:      exe.stateCache,
			NameReg:    exe.nameRegCache,
			Logger:     exe.logger,
		},
		payload.TypePermissions: &contexts.PermissionsContext{
			State:  exe.stateCache,
			Logger: exe.logger,
		},
		payload.TypeGovernance: &contexts.GovernanceContext{
			ValidatorSet: exe.validatorCache,
			State:        exe.stateCache,
			Logger:       exe.logger,
		},
		payload.TypeBond: &contexts.BondContext{
			ValidatorSet: exe.validatorCache,
			State:        exe.stateCache,
			Logger:       exe.logger,
		},
		payload.TypeUnbond: &contexts.UnbondContext{
			ValidatorSet: exe.validatorCache,
			State:        exe.stateCache,
			Logger:       exe.logger,
		},
		payload.TypeIdentify: &contexts.IdentifyContext{
			NodeWriter:  exe.nodeRegCache,
			StateReader: exe.stateCache,
			Logger:      exe.logger,
		},
	}

	exe.contexts = map[payload.Type]contexts.Context{
		payload.TypeProposal: &contexts.ProposalContext{
			ChainID:           params.ChainID,
			ProposalThreshold: params.ProposalThreshold,
			State:             exe.stateCache,
			ProposalReg:       exe.proposalRegCache,
			Logger:            exe.logger,
			Contexts:          baseContexts,
		},
	}

	// Copy over base contexts
	for k, v := range baseContexts {
		exe.contexts[k] = v
	}

	return exe, nil
}

func (exe *executor) AddContext(ty payload.Type, ctx contexts.Context) *executor {
	exe.contexts[ty] = ctx
	return exe
}

// If the tx is invalid, an error will be returned.
// Unlike ExecBlock(), state will not be altered.
func (exe *executor) Execute(txEnv *txs.Envelope) (txe *exec.TxExecution, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in executor.Execute(%s): %v\n%s", txEnv.String(), r,
				debug.Stack())
		}
	}()

	logger := exe.logger.WithScope("executor.Execute(tx txs.Tx)").With(
		"height", exe.block.Height,
		"run_call", exe.runCall,
		structure.TxHashKey, txEnv.Tx.Hash())

	logger.InfoMsg("Executing transaction", "tx", txEnv.String())

	// Verify transaction signature against inputs
	err = txEnv.Verify(exe.params.ChainID)
	if err != nil {
		logger.InfoMsg("Transaction Verify failed", structure.ErrorKey, err)
		return nil, err
	}

	if txExecutor, ok := exe.contexts[txEnv.Tx.Type()]; ok {
		// Establish new TxExecution
		txe := exe.block.Tx(txEnv)
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("recovered from panic in executor.Execute(%s): %v\n%s", txEnv.String(), r,
					debug.Stack())
			}
		}()

		err = exe.validateInputsAndStorePublicKeys(txEnv)
		if err != nil {
			logger.InfoMsg("Transaction validate failed", structure.ErrorKey, err)
			txe.PushError(err)
			return nil, err
		}

		err = txExecutor.Execute(txe, txe.Envelope.Tx.Payload)
		if err != nil {
			logger.InfoMsg("Transaction execution failed", structure.ErrorKey, err)
			txe.PushError(err)
			return nil, err
		}

		// Increment sequence numbers for Tx inputs
		err = exe.updateSequenceNumbers(txEnv)
		if err != nil {
			logger.InfoMsg("Updating sequences failed", structure.ErrorKey, err)
			txe.PushError(err)
			return nil, err
		}
		// Return execution for this tx
		return txe, nil
	}
	return nil, fmt.Errorf("unknown transaction type: %v", txEnv.Tx.Type())
}

// Validate inputs, check sequence numbers and capture public keys
func (exe *executor) validateInputsAndStorePublicKeys(txEnv *txs.Envelope) error {
	for s, in := range txEnv.Tx.GetInputs() {
		err := exe.updateSignatory(txEnv.Signatories[s])
		if err != nil {
			return fmt.Errorf("failed to update public key for input %X: %v", in.Address, err)
		}
		acc, err := exe.stateCache.GetAccount(in.Address)
		if err != nil {
			return err
		}
		if acc == nil {
			return fmt.Errorf("validateInputs() expects to be able to retrieve account %v but it was not found",
				in.Address)
		}
		if in.Address != acc.GetAddress() {
			return fmt.Errorf("trying to validate input from address %v but passed account %v", in.Address,
				acc.GetAddress())
		}
		// Check sequences
		if acc.Sequence+1 != in.Sequence {
			return errors.Errorf(errors.Codes.InvalidSequence, "Error invalid sequence in input %v: input has sequence %d, but account has sequence %d, "+
				"so expected input to have sequence %d", in, in.Sequence, acc.Sequence, acc.Sequence+1)
		}
		// Check amount
		if txEnv.Tx.Type() != payload.TypeUnbond && acc.Balance < in.Amount {
			return errors.Codes.InsufficientFunds
		}
		// Check for Input permission
		globalPerms, err := acmstate.GlobalAccountPermissions(exe.stateCache)
		if err != nil {
			return err
		}
		v, err := acc.Permissions.Base.Compose(globalPerms.Base).Get(permission.Input)
		if err != nil {
			return err
		}
		if !v {
			return errors.Codes.NoInputPermission
		}
	}
	return nil
}

func (exe *executor) updateSignatory(sig txs.Signatory) error {
	// pointer dereferences are safe since txEnv.Validate() is run by
	// txEnv.Verify() above which checks they are non-nil
	acc, err := exe.stateCache.GetAccount(*sig.Address)
	if err != nil {
		return fmt.Errorf("error getting account on which to set public key: %v", *sig.Address)
	} else if acc == nil {
		return fmt.Errorf("account %s does not exist", sig.Address)
	}
	// Important that verify has been run against signatories at this point
	if sig.PublicKey.GetAddress() != acc.Address {
		return fmt.Errorf("unexpected mismatch between address %v and supplied public key %v",
			acc.Address, sig.PublicKey)
	}
	acc.PublicKey = *sig.PublicKey
	return exe.stateCache.UpdateAccount(acc)
}

// Commit the current state - optionally pass in the tendermint ABCI header for that to be included with the BeginBlock
// StreamEvent
func (exe *executor) Commit(header *types.Header) (stateHash []byte, err error) {
	// The write lock to the executor is controlled by the caller (e.g. abci.App) so we do not acquire it here to avoid
	// deadlock
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in executor.Commit(): %v\n%s", r, debug.Stack())
		}
	}()
	// Capture height
	height := exe.block.Height
	exe.logger.InfoMsg("Executor committing", "height", exe.block.Height)
	// Form BlockExecution for this block from TxExecutions and Tendermint block header
	blockExecution, err := exe.finaliseBlockExecution(header)
	if err != nil {
		return nil, err
	}
	// First commit the app state, this app hash will not get checkpointed until the next block when we are sure
	// that nothing in the downstream commit process could have failed. At worst we go back one block.
	hash, version, err := exe.state.Update(func(ws state.Updatable) error {
		// flush the caches
		err := exe.stateCache.Sync(ws)
		if err != nil {
			return err
		}
		err = exe.metadataCache.Sync(ws)
		if err != nil {
			return err
		}
		err = exe.nameRegCache.Sync(ws)
		if err != nil {
			return err
		}
		err = exe.nodeRegCache.Sync(ws)
		if err != nil {
			return err
		}
		err = exe.proposalRegCache.Sync(ws)
		if err != nil {
			return err
		}
		err = exe.validatorCache.Sync(ws)
		if err != nil {
			return err
		}
		err = ws.AddBlock(blockExecution)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Complete flushing of caches by resetting them to the state we have just committed
	err = exe.Reset()
	if err != nil {
		return nil, err
	}

	expectedHeight := HeightAtVersion(version)
	if expectedHeight != height {
		return nil, fmt.Errorf("expected height at state tree version %d is %d but actual height is %d",
			version, expectedHeight, height)
	}
	// Now state is fully committed publish events (this should be the last thing we do)
	exe.publishBlock(blockExecution)
	return hash, nil
}

func (exe *executor) Reset() error {
	// As with Commit() we do not take the write lock here
	exe.stateCache.Reset(exe.state)
	exe.metadataCache.Reset(exe.state)
	exe.nameRegCache.Reset(exe.state)
	exe.nodeRegCache.Reset(exe.state)
	exe.proposalRegCache.Reset(exe.state)
	exe.validatorCache.Reset(exe.state)
	return nil
}

// executor exposes access to the underlying state cache protected by a RWMutex that prevents access while locked
// (during an ABCI commit). while access can occur (and needs to continue for CheckTx/DeliverTx to make progress)
// through calls to Execute() external readers will be blocked until the executor is unlocked that allows the Transactor
// to always access the freshest mempool state as needed by accounts.SequentialSigningAccount
//
// Accounts
func (exe *executor) GetAccount(address crypto.Address) (*acm.Account, error) {
	exe.RLock()
	defer exe.RUnlock()
	return exe.stateCache.GetAccount(address)
}

// Storage
func (exe *executor) GetStorage(address crypto.Address, key binary.Word256) ([]byte, error) {
	exe.RLock()
	defer exe.RUnlock()
	return exe.stateCache.GetStorage(address, key)
}

func (exe *executor) PendingValidators() validator.IterableReader {
	return exe.validatorCache.Delta
}

func (exe *executor) finaliseBlockExecution(header *types.Header) (*exec.BlockExecution, error) {
	if header != nil && uint64(header.Height) != exe.block.Height {
		return nil, fmt.Errorf("trying to finalise block execution with height %v but passed Tendermint"+
			"block header at height %v", exe.block.Height, header.Height)
	}
	// Capture BlockExecution to return
	be := exe.block
	// Set the header when provided
	be.Header = header
	// My default the predecessor of the next block is the is the predecessor of the current block
	// (in case the current block has no transactions - since we do not currently store empty blocks in state, see
	// /execution/state/events.go)
	predecessor := be.PredecessorHeight
	if len(be.TxExecutions) > 0 {
		// If the current block has transactions then it will be the predecessor of the next block
		predecessor = be.Height
	}
	// Start new execution for the next height
	exe.block = &exec.BlockExecution{
		Height:            exe.block.Height + 1,
		PredecessorHeight: predecessor,
	}
	return be, nil
}

// update sequence numbers
func (exe *executor) updateSequenceNumbers(txEnv *txs.Envelope) error {
	for _, sig := range txEnv.Signatories {
		acc, err := exe.stateCache.GetAccount(*sig.Address)
		if err != nil {
			return fmt.Errorf("error getting account on which to set public key: %v", *sig.Address)
		}

		exe.logger.TraceMsg("Incrementing sequence number Tx signatory/input",
			"height", exe.block.Height,
			"tag", "sequence",
			"account", acc.Address,
			"old_sequence", acc.Sequence,
			"new_sequence", acc.Sequence+1)

		acc.Sequence++
		err = exe.stateCache.UpdateAccount(acc)
		if err != nil {
			return fmt.Errorf("error updating account after incrementing sequence: %v", err)
		}
	}
	return nil
}

func (exe *executor) publishBlock(blockExecution *exec.BlockExecution) {
	for _, txe := range blockExecution.TxExecutions {
		publishErr := exe.emitter.Publish(context.Background(), txe, txe)
		if publishErr != nil {
			exe.logger.InfoMsg("Error publishing TxExecution",
				"height", blockExecution.Height,
				structure.TxHashKey, txe.TxHash,
				structure.ErrorKey, publishErr)
		}
	}
	publishErr := exe.emitter.Publish(context.Background(), blockExecution, blockExecution)
	if publishErr != nil {
		exe.logger.InfoMsg("Error publishing BlockExecution",
			"height", blockExecution.Height,
			structure.ErrorKey, publishErr)
	}
}
