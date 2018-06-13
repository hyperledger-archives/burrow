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
	"fmt"
	"runtime/debug"
	"sync"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/binary"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/executors"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
)

type Executor interface {
	Execute(txEnv *txs.Envelope) error
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
	Commit() (stateHash []byte, err error)
}

type executor struct {
	sync.RWMutex
	runCall      bool
	state        *State
	stateCache   state.Cache
	nameRegCache *names.Cache
	eventCache   *event.Cache
	logger       *logging.Logger
	vmOptions    []func(*evm.VM)
	txExecutors  map[payload.Type]Executor
}

var _ BatchExecutor = (*executor)(nil)

// Wraps a cache of what is variously known as the 'check cache' and 'mempool'
func NewBatchChecker(backend *State, tip *bcm.Tip, logger *logging.Logger,
	options ...ExecutionOption) BatchExecutor {

	return newExecutor("CheckCache", false, backend, tip, event.NewNoOpPublisher(),
		logger.WithScope("NewBatchExecutor"), options...)
}

func NewBatchCommitter(backend *State, tip *bcm.Tip, publisher event.Publisher, logger *logging.Logger,
	options ...ExecutionOption) BatchCommitter {

	return newExecutor("CommitCache", true, backend, tip, publisher,
		logger.WithScope("NewBatchCommitter"), options...)
}

func newExecutor(name string, runCall bool, backend *State, tip *bcm.Tip, publisher event.Publisher,
	logger *logging.Logger, options ...ExecutionOption) *executor {

	exe := &executor{
		runCall:      runCall,
		state:        backend,
		stateCache:   state.NewCache(backend, state.Name(name)),
		eventCache:   event.NewEventCache(publisher),
		nameRegCache: names.NewCache(backend),
		logger:       logger.With(structure.ComponentKey, "Executor"),
	}
	for _, option := range options {
		option(exe)
	}
	exe.txExecutors = map[payload.Type]Executor{
		payload.TypeSend: &executors.SendContext{
			StateWriter:    exe.stateCache,
			EventPublisher: exe.eventCache,
			Logger:         exe.logger,
		},
		payload.TypeCall: &executors.CallContext{
			StateWriter:    exe.stateCache,
			EventPublisher: exe.eventCache,
			Tip:            tip,
			RunCall:        runCall,
			VMOptions:      exe.vmOptions,
			Logger:         exe.logger,
		},
		payload.TypeName: &executors.NameContext{
			StateWriter:    exe.stateCache,
			EventPublisher: exe.eventCache,
			NameReg:        exe.nameRegCache,
			Tip:            tip,
			Logger:         exe.logger,
		},
		payload.TypePermissions: &executors.PermissionsContext{
			StateWriter:    exe.stateCache,
			EventPublisher: exe.eventCache,
			Logger:         exe.logger,
		},
	}
	return exe
}

// If the tx is invalid, an error will be returned.
// Unlike ExecBlock(), state will not be altered.
func (exe *executor) Execute(txEnv *txs.Envelope) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in executor.Execute(%s): %v\n%s", txEnv.String(), r,
				debug.Stack())
		}
	}()

	logger := exe.logger.WithScope("executor.Execute(tx txs.Tx)").With(
		"run_call", exe.runCall,
		"tx_hash", txEnv.Tx.Hash())

	logger.TraceMsg("Executing transaction", "tx", txEnv.String())

	// Verify transaction signature against inputs
	err = txEnv.Verify(exe.stateCache)
	if err != nil {
		return err
	}

	if txExecutor, ok := exe.txExecutors[txEnv.Tx.Type()]; ok {
		return txExecutor.Execute(txEnv)
	}
	return fmt.Errorf("unknown transaction type: %v", txEnv.Tx.Type())
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

func (exe *executor) Commit() (hash []byte, err error) {
	// The write lock to the executor is controlled by the caller (e.g. abci.App) so we do not acquire it here to avoid
	// deadlock
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in executor.Commit(): %v\n%v", r, debug.Stack())
		}
	}()
	// flush the caches
	err = exe.stateCache.Flush(exe.state)
	if err != nil {
		return nil, err
	}
	err = exe.nameRegCache.Flush(exe.state)
	if err != nil {
		return nil, err
	}
	// save state to disk
	err = exe.state.Save()
	if err != nil {
		return nil, err
	}
	// flush events to listeners
	defer exe.eventCache.Flush()
	return exe.state.Hash(), nil
}

func (exe *executor) Reset() error {
	// As with Commit() we do not take the write lock here
	exe.stateCache.Reset(exe.state)
	exe.nameRegCache.Reset(exe.state)
	return nil
}
