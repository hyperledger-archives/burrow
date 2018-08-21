// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package execution

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/contexts"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

type Executor interface {
	Execute(txEnv *txs.Envelope) (*exec.TxExecution, error)
}

type Context interface {
	Execute(txe *exec.TxExecution) error
}

type ExecutorState interface {
	Update(updater func(ws Updatable) error) (hash []byte, err error)
	names.Reader
	state.IterableReader
}

type BatchExecutor interface {
	// Provides access to write lock for a BatchExecutor so reads can be prevented for the duration of a commit
	sync.Locker
	state.Reader
	// Execute transaction against block cache (i.e. block buffer)
	Executor
	// Reset executor to underlying State
	Reset() error
}

// Executes transactions
type BatchCommitter interface {
	BatchExecutor
	// Commit execution results to underlying State and provide opportunity
	// to mutate state before it is saved
	Commit(blockHash []byte, blockTime time.Time, header *abciTypes.Header) (stateHash []byte, err error)
}

type executor struct {
	sync.RWMutex
	runCall        bool
	blockchain     *bcm.Blockchain
	state          ExecutorState
	stateCache     *state.Cache
	nameRegCache   *names.Cache
	publisher      event.Publisher
	blockExecution *exec.BlockExecution
	logger         *logging.Logger
	vmOptions      []func(*evm.VM)
	contexts       map[payload.Type]Context
}

var _ BatchExecutor = (*executor)(nil)

// Wraps a cache of what is variously known as the 'check cache' and 'mempool'
func NewBatchChecker(backend ExecutorState, blockchain *bcm.Blockchain, logger *logging.Logger,
	options ...ExecutionOption) BatchExecutor {

	exe := newExecutor("CheckCache", false, backend, blockchain, event.NewNoOpPublisher(),
		logger.WithScope("NewBatchExecutor"), options...)

	return exe.AddContext(payload.TypeGovernance,
		&contexts.GovernanceContext{
			ValidatorSet: exe.blockchain.ValidatorChecker(),
			StateWriter:  exe.stateCache,
			Logger:       exe.logger,
		},
	)
}

func NewBatchCommitter(backend ExecutorState, blockchain *bcm.Blockchain, emitter event.Publisher,
	logger *logging.Logger, options ...ExecutionOption) BatchCommitter {

	exe := newExecutor("CommitCache", true, backend, blockchain, emitter,
		logger.WithScope("NewBatchCommitter"), options...)

	return exe.AddContext(payload.TypeGovernance,
		&contexts.GovernanceContext{
			ValidatorSet: exe.blockchain.ValidatorWriter(),
			StateWriter:  exe.stateCache,
			Logger:       exe.logger,
		},
	)
}

func newExecutor(name string, runCall bool, backend ExecutorState, blockchain *bcm.Blockchain, publisher event.Publisher,
	logger *logging.Logger, options ...ExecutionOption) *executor {
	exe := &executor{
		runCall:      runCall,
		state:        backend,
		blockchain:   blockchain,
		stateCache:   state.NewCache(backend, state.Name(name)),
		nameRegCache: names.NewCache(backend),
		publisher:    publisher,
		blockExecution: &exec.BlockExecution{
			Height: blockchain.LastBlockHeight() + 1,
		},
		logger: logger.With(structure.ComponentKey, "Executor"),
	}
	for _, option := range options {
		option(exe)
	}
	exe.contexts = map[payload.Type]Context{
		payload.TypeSend: &contexts.SendContext{
			Tip:         blockchain,
			StateWriter: exe.stateCache,
			Logger:      exe.logger,
		},
		payload.TypeCall: &contexts.CallContext{
			Tip:         blockchain,
			StateWriter: exe.stateCache,
			RunCall:     runCall,
			VMOptions:   exe.vmOptions,
			Logger:      exe.logger,
		},
		payload.TypeName: &contexts.NameContext{
			Tip:         blockchain,
			StateWriter: exe.stateCache,
			NameReg:     exe.nameRegCache,
			Logger:      exe.logger,
		},
		payload.TypePermissions: &contexts.PermissionsContext{
			Tip:         blockchain,
			StateWriter: exe.stateCache,
			Logger:      exe.logger,
		},
	}
	return exe
}

func (exe *executor) AddContext(ty payload.Type, ctx Context) *executor {
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
		"run_call", exe.runCall,
		"tx_hash", txEnv.Tx.Hash())

	logger.InfoMsg("Executing transaction", "tx", txEnv.String())

	// Verify transaction signature against inputs
	err = txEnv.Verify(exe.stateCache, exe.blockchain.ChainID())
	if err != nil {
		logger.InfoMsg("Transaction Verify failed", structure.ErrorKey, err)
		return nil, err
	}

	if txExecutor, ok := exe.contexts[txEnv.Tx.Type()]; ok {
		// Establish new TxExecution
		txe := exe.blockExecution.Tx(txEnv)
		// Validate inputs and check sequence numbers
		err = txEnv.Tx.ValidateInputs(exe.stateCache)
		if err != nil {
			logger.InfoMsg("Transaction validate failed", structure.ErrorKey, err)
			return nil, err
		}
		err = txExecutor.Execute(txe)
		if err != nil {
			logger.InfoMsg("Transaction execution failed", structure.ErrorKey, err)
			return nil, err
		}
		// Initialise public keys and increment sequence numbers for Tx inputs
		err = exe.updateSignatories(txEnv)
		if err != nil {
			logger.InfoMsg("Updating signatories failed", structure.ErrorKey, err)
			return nil, err
		}
		// Return execution for this tx
		return txe, nil
	}
	return nil, fmt.Errorf("unknown transaction type: %v", txEnv.Tx.Type())
}

func (exe *executor) Commit(blockHash []byte, blockTime time.Time, header *abciTypes.Header) (_ []byte, err error) {

	// The write lock to the executor is controlled by the caller (e.g. abci.App) so we do not acquire it here to avoid
	// deadlock
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in executor.Commit(): %v\n%s", r, debug.Stack())
		}
	}()
	exe.logger.InfoMsg("Executor committing", "height", exe.blockExecution.Height)
	// Form BlockExecution for this block from TxExecutions and Tendermint block header
	blockExecution, err := exe.finaliseBlockExecution(header)
	if err != nil {
		return nil, err
	}

	// First commit the app state, this app hash will not get checkpointed until the next block when we are sure
	// that nothing in the downstream commit process could have failed. At worst we go back one block.
	hash, err := exe.state.Update(func(ws Updatable) error {
		// flush the caches
		err := exe.stateCache.Flush(ws, exe.state)
		if err != nil {
			return err
		}
		err = exe.nameRegCache.Flush(ws, exe.state)
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
	// Now state is committed publish events
	for _, txe := range blockExecution.TxExecutions {
		publishErr := exe.publisher.Publish(context.Background(), txe, txe.Tagged())
		if publishErr != nil {
			exe.logger.InfoMsg("Error publishing TxExecution",
				"tx_hash", txe.TxHash,
				structure.ErrorKey, publishErr)
		}
	}
	publishErr := exe.publisher.Publish(context.Background(), blockExecution, blockExecution.Tagged())
	if publishErr != nil {
		exe.logger.InfoMsg("Error publishing BlockExecution",
			"height", blockExecution.Height, structure.ErrorKey, publishErr)
	}
	// Commit to our blockchain state which will checkpoint the previous app hash by saving it to the database
	// (we know the previous app hash is safely committed because we are about to commit the next)
	totalPowerChange, totalFlow, err := exe.blockchain.CommitBlock(blockTime, blockHash, hash)
	if err != nil {
		panic(fmt.Errorf("could not commit block to blockchain state: %v", err))
	}
	exe.logger.InfoMsg("Committed block",
		"total_validator_power", exe.blockchain.CurrentValidators().TotalPower(),
		"total_validator_power_change", totalPowerChange,
		"total_validator_flow", totalFlow)

	return hash, nil
}

func (exe *executor) Reset() error {
	// As with Commit() we do not take the write lock here
	exe.stateCache.Reset(exe.state)
	exe.nameRegCache.Reset(exe.state)
	return nil
}

// executor exposes access to the underlying state cache protected by a RWMutex that prevents access while locked
// (during an ABCI commit). while access can occur (and needs to continue for CheckTx/DeliverTx to make progress)
// through calls to Execute() external readers will be blocked until the executor is unlocked that allows the Transactor
// to always access the freshest mempool state as needed by accounts.SequentialSigningAccount
//
// Accounts
func (exe *executor) GetAccount(address crypto.Address) (acm.Account, error) {
	exe.RLock()
	defer exe.RUnlock()
	return exe.stateCache.GetAccount(address)
}

// Storage
func (exe *executor) GetStorage(address crypto.Address, key binary.Word256) (binary.Word256, error) {
	exe.RLock()
	defer exe.RUnlock()
	return exe.stateCache.GetStorage(address, key)
}

func (exe *executor) finaliseBlockExecution(header *abciTypes.Header) (*exec.BlockExecution, error) {
	if header != nil && uint64(header.Height) != exe.blockExecution.Height {
		return nil, fmt.Errorf("trying to finalise block execution with height %v but passed Tendermint"+
			"block header at height %v", exe.blockExecution.Height, header.Height)
	}
	// Capture BlockExecution to return
	be := exe.blockExecution
	// Set the header when provided
	be.BlockHeader = exec.BlockHeaderFromHeader(header)
	// Start new execution for the next height
	exe.blockExecution = &exec.BlockExecution{
		Height: exe.blockExecution.Height + 1,
	}
	return be, nil
}

// Capture public keys and update sequence numbers
func (exe *executor) updateSignatories(txEnv *txs.Envelope) error {
	for _, sig := range txEnv.Signatories {
		// pointer dereferences are safe since txEnv.Validate() is run by txEnv.Verify() above which checks they are
		// non-nil
		acc, err := state.GetMutableAccount(exe.stateCache, *sig.Address)
		if err != nil {
			return fmt.Errorf("error getting account on which to set public key: %v", *sig.Address)
		}
		// Important that verify has been run against signatories at this point
		if sig.PublicKey.Address() != acc.Address() {
			return fmt.Errorf("unexpected mismatch between address %v and supplied public key %v",
				acc.Address(), sig.PublicKey)
		}
		acc.SetPublicKey(*sig.PublicKey)

		exe.logger.TraceMsg("Incrementing sequence number Tx signatory/input",
			"tag", "sequence",
			"account", acc.Address(),
			"old_sequence", acc.Sequence(),
			"new_sequence", acc.Sequence()+1)
		acc.IncSequence()
		err = exe.stateCache.UpdateAccount(acc)
		if err != nil {
			return fmt.Errorf("error updating account after setting public key: %v", err)
		}
	}
	return nil
}
