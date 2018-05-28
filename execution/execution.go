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
	"errors"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/hyperledger/burrow/finterra"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/binary"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/hyperledger/burrow/txs"
	rpc_core "github.com/tendermint/tendermint/rpc/core"
)

// TODO: make configurable
const GasLimit = uint64(1000000)

type BatchExecutor interface {
	// Provides access to write lock for a BatchExecutor so reads can be prevented for the duration of a commit
	sync.Locker
	state.Reader
	// Execute transaction against block cache (i.e. block buffer)
	Execute(tx txs.Tx) error
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
	chainID      string
	tip          bcm.Tip
	runCall      bool
	state        *State
	stateCache   state.Cache
	nameRegCache *NameRegCache
	publisher    event.Publisher
	eventCache   *event.Cache
	logger       *logging.Logger
	vmOptions    []func(*evm.VM)
	valPoolCache *finterra.ValidatorPoolCache
}

var _ BatchExecutor = (*executor)(nil)

// Wraps a cache of what is variously known as the 'check cache' and 'mempool'
func NewBatchChecker(backend *State, chainID string, tip bcm.Tip, logger *logging.Logger,
	options ...ExecutionOption) BatchExecutor {

	return newExecutor("CheckCache", false, backend, chainID, tip, event.NewNoOpPublisher(),
		logger.WithScope("NewBatchExecutor"), options...)
}

func NewBatchCommitter(backend *State, chainID string, tip bcm.Tip, publisher event.Publisher, logger *logging.Logger,
	options ...ExecutionOption) BatchCommitter {

	return newExecutor("CommitCache", true, backend, chainID, tip, publisher,
		logger.WithScope("NewBatchCommitter"), options...)
}

func newExecutor(name string, runCall bool, backend *State, chainID string, tip bcm.Tip, publisher event.Publisher,
	logger *logging.Logger, options ...ExecutionOption) *executor {

	exe := &executor{
		chainID:      chainID,
		tip:          tip,
		runCall:      runCall,
		state:        backend,
		stateCache:   state.NewCache(backend, state.Name(name)),
		nameRegCache: NewNameRegCache(backend),
		publisher:    publisher,
		eventCache:   event.NewEventCache(publisher),
		logger:       logger.With(structure.ComponentKey, "Executor"),
		valPoolCache: finterra.NewValidatorPoolCache(backend),
	}
	for _, option := range options {
		option(exe)
	}
	return exe
}

// executor exposes access to the underlying state cache protected by a RWMutex that prevents access while locked
// (during an ABCI commit). while access can occur (and needs to continue for CheckTx/DeliverTx to make progress)
// through calls to Execute() external readers will be blocked until the executor is unlocked that allows the Transactor
// to always access the freshest mempool state as needed by accounts.SequentialSigningAccount
//
// Accounts
func (exe *executor) GetAccount(address acm.Address) (acm.Account, error) {
	exe.RLock()
	defer exe.RUnlock()
	return exe.stateCache.GetAccount(address)
}

// Storage
func (exe *executor) GetStorage(address acm.Address, key binary.Word256) (binary.Word256, error) {
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
	/// update validator-set
	err = exe.valPoolCache.Flush(exe.state, exe.tip.ValidatorSet())
	if err != nil {
		return nil, err
	}
	exe.valPoolCache.Reset()
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
	exe.valPoolCache.Reset() /// TODO: MOSTAFA: do we need to pass ValidatorPoolGetter again for reset object!
	return nil
}

// If the tx is invalid, an error will be returned.
// Unlike ExecBlock(), state will not be altered.
func (exe *executor) Execute(tx txs.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in executor.Execute(%s): %v\n%s", tx.String(), r,
				debug.Stack())
		}
	}()

	txHash := tx.Hash(exe.chainID)
	logger := exe.logger.WithScope("executor.Execute(tx txs.Tx)").With(
		"run_call", exe.runCall,
		"tx_hash", txHash)
	logger.TraceMsg("Executing transaction", "tx", tx.String())
	// TODO: do something with fees
	fees := uint64(0)

	// Exec tx
	switch tx := tx.(type) {
	case *txs.SendTx:
		accounts, err := getInputAccounts(exe.stateCache, tx.Inputs)
		if err != nil {
			return err
		}

		// ensure all inputs have send permissions
		if !hasSendPermission(exe.stateCache, accounts, logger) {
			return fmt.Errorf("at least one input lacks permission for SendTx")
		}

		// add outputs to accounts map
		// if any outputs don't exist, all inputs must have CreateAccount perm
		accounts, err = getOrMakeOutputs(exe.stateCache, accounts, tx.Outputs, logger)
		if err != nil {
			return err
		}

		signBytes := acm.SignBytes(exe.chainID, tx)
		inTotal, err := validateInputs(accounts, signBytes, tx.Inputs)
		if err != nil {
			return err
		}
		outTotal, err := validateOutputs(tx.Outputs)
		if err != nil {
			return err
		}
		if outTotal > inTotal {
			return txs.ErrTxInsufficientFunds
		}
		fee := inTotal - outTotal
		fees += fee

		// Good! Adjust accounts
		err = adjustByInputs(accounts, tx.Inputs, logger)
		if err != nil {
			return err
		}

		err = adjustByOutputs(accounts, tx.Outputs)
		if err != nil {
			return err
		}

		for _, acc := range accounts {
			exe.stateCache.UpdateAccount(acc)
		}

		if exe.eventCache != nil {
			for _, input := range tx.Inputs {
				events.PublishAccountInput(exe.eventCache, input.Address, txHash, tx, nil, "")
			}

			for _, output := range tx.Outputs {
				events.PublishAccountOutput(exe.eventCache, output.Address, txHash, tx, nil, "")
			}
		}
		return nil

	case *txs.CallTx:
		var inAcc acm.MutableAccount
		var outAcc acm.Account

		// Validate input
		inAcc, err := state.GetMutableAccount(exe.stateCache, tx.Input.Address)
		if err != nil {
			return err
		}
		if inAcc == nil {
			logger.InfoMsg("Cannot find input account",
				"tx_input", tx.Input)
			return txs.ErrTxInvalidAddress
		}

		// Calling a nil destination is defined as requesting contract creation
		createContract := tx.Address == nil
		if createContract {
			if !hasCreateContractPermission(exe.stateCache, inAcc, logger) {
				return fmt.Errorf("account %s does not have CreateContract permission", tx.Input.Address)
			}
		} else {
			if !hasCallPermission(exe.stateCache, inAcc, logger) {
				return fmt.Errorf("account %s does not have Call permission", tx.Input.Address)
			}
		}

		// pubKey should be present in either "inAcc" or "tx.Input"
		if err := checkInputPubKey(inAcc, tx.Input); err != nil {
			logger.InfoMsg("Cannot find public key for input account",
				"tx_input", tx.Input)
			return err
		}
		signBytes := acm.SignBytes(exe.chainID, tx)
		amount, err := validateInput(inAcc, signBytes, tx.Input)
		if err != nil {
			logger.InfoMsg("validateInput failed",
				"tx_input", tx.Input, structure.ErrorKey, err)
			return err
		}
		if tx.Input.Amount < tx.Fee {
			logger.InfoMsg("Sender did not send enough to cover the fee",
				"tx_input", tx.Input)
			return txs.ErrTxInsufficientFunds
		}

		if !createContract {
			// check if its a native contract
			if evm.RegisteredNativeContract(tx.Address.Word256()) {
				return fmt.Errorf("attempt to call a native contract at %s, "+
					"but native contracts cannot be called using CallTx. Use a "+
					"contract that calls the native contract or the appropriate tx "+
					"type (eg. PermissionsTx, NameTx)", tx.Address)
			}

			// Output account may be nil if we are still in mempool and contract was created in same block as this tx
			// but that's fine, because the account will be created properly when the create tx runs in the block
			// and then this won't return nil. otherwise, we take their fee
			// Note: tx.Address == nil iff createContract so dereference is okay
			outAcc, err = exe.stateCache.GetAccount(*tx.Address)
			if err != nil {
				return err
			}
		}

		logger.Trace.Log("output_account", outAcc)

		// Good!
		value := amount - tx.Fee

		logger.TraceMsg("Incrementing sequence number for CallTx",
			"tag", "sequence",
			"account", inAcc.Address(),
			"old_sequence", inAcc.Sequence(),
			"new_sequence", inAcc.Sequence()+1)

		inAcc.IncSequence()
		inAcc.SubtractFromBalance(tx.Fee)
		if err != nil {
			return err
		}

		exe.stateCache.UpdateAccount(inAcc)

		// The logic in runCall MUST NOT return.
		if exe.runCall {
			// VM call variables
			var (
				gas     uint64             = tx.GasLimit
				err     error              = nil
				caller  acm.MutableAccount = acm.AsMutableAccount(inAcc)
				callee  acm.MutableAccount = nil // initialized below
				code    []byte             = nil
				ret     []byte             = nil
				txCache                    = state.NewCache(exe.stateCache, state.Name("TxCache"))
				params                     = evm.Params{
					BlockHeight: exe.tip.LastBlockHeight(),
					BlockHash:   binary.LeftPadWord256(exe.tip.LastBlockHash()),
					BlockTime:   exe.tip.LastBlockTime().Unix(),
					GasLimit:    GasLimit,
				}
			)

			if !createContract && (outAcc == nil || len(outAcc.Code()) == 0) {
				// if you call an account that doesn't exist
				// or an account with no code then we take fees (sorry pal)
				// NOTE: it's fine to create a contract and call it within one
				// block (sequence number will prevent re-ordering of those txs)
				// but to create with one contract and call with another
				// you have to wait a block to avoid a re-ordering attack
				// that will take your fees
				if outAcc == nil {
					logger.InfoMsg("Call to address that does not exist",
						"caller_address", inAcc.Address(),
						"callee_address", tx.Address)
				} else {
					logger.InfoMsg("Call to address that holds no code",
						"caller_address", inAcc.Address(),
						"callee_address", tx.Address)
				}
				err = txs.ErrTxInvalidAddress
				goto CALL_COMPLETE
			}

			// get or create callee
			if createContract {
				// We already checked for permission
				callee = evm.DeriveNewAccount(caller, state.GlobalAccountPermissions(exe.state),
					logger.With(
						"tx", tx.String(),
						"tx_hash", txHash,
						"run_call", exe.runCall,
					))
				code = tx.Data
				logger.TraceMsg("Creating new contract",
					"contract_address", callee.Address(),
					"init_code", code)
			} else {
				callee = acm.AsMutableAccount(outAcc)
				code = callee.Code()
				logger.TraceMsg("Calling existing contract",
					"contract_address", callee.Address(),
					"input", tx.Data,
					"contract_code", code)
			}
			logger.Trace.Log("callee", callee.Address().String())

			// Run VM call and sync txCache to exe. .
			{ // Capture scope for goto.
				// Write caller/callee to txCache.
				txCache.UpdateAccount(caller)
				txCache.UpdateAccount(callee)
				vmach := evm.NewVM(txCache, params, caller.Address(), tx.Hash(exe.chainID), logger, exe.vmOptions...)
				vmach.SetPublisher(exe.eventCache)
				// NOTE: Call() transfers the value from caller to callee iff call succeeds.
				ret, err = vmach.Call(caller, callee, code, tx.Data, value, &gas)
				if err != nil {
					// Failure. Charge the gas fee. The 'value' was otherwise not transferred.
					logger.InfoMsg("Error on execution",
						structure.ErrorKey, err)
					goto CALL_COMPLETE
				}

				logger.TraceMsg("Successful execution")
				if createContract {
					callee.SetCode(ret)
				}
				txCache.Sync(exe.stateCache)
			}

		CALL_COMPLETE: // err may or may not be nil.

			// Create a receipt from the ret and whether it erred.
			logger.TraceMsg("VM call complete",
				"caller", caller,
				"callee", callee,
				"return", ret,
				structure.ErrorKey, err)

			// Fire Events for sender and receiver
			// a separate event will be fired from vm for each additional call
			if exe.eventCache != nil {
				exception := ""
				if err != nil {
					exception = err.Error()
				}
				txHash := tx.Hash(exe.chainID)
				events.PublishAccountInput(exe.eventCache, tx.Input.Address, txHash, tx, ret, exception)
				if tx.Address != nil {
					events.PublishAccountOutput(exe.eventCache, *tx.Address, txHash, tx, ret, exception)
				}
			}
		} else {
			// The mempool does not call txs until
			// the proposer determines the order of txs.
			// So mempool will skip the actual .Call(),
			// and only deduct from the caller's balance.
			err = inAcc.SubtractFromBalance(value)
			if err != nil {
				return err
			}
			if createContract {
				// This is done by DeriveNewAccount when runCall == true
				logger.TraceMsg("Incrementing sequence number since creates contract",
					"tag", "sequence",
					"account", inAcc.Address(),
					"old_sequence", inAcc.Sequence(),
					"new_sequence", inAcc.Sequence()+1)
				inAcc.IncSequence()
			}
			exe.stateCache.UpdateAccount(inAcc)
		}

		return nil

	case *txs.NameTx:
		// Validate input
		inAcc, err := state.GetMutableAccount(exe.stateCache, tx.Input.Address)
		if err != nil {
			return err
		}
		if inAcc == nil {
			logger.InfoMsg("Cannot find input account",
				"tx_input", tx.Input)
			return txs.ErrTxInvalidAddress
		}
		// check permission
		if !hasNamePermission(exe.stateCache, inAcc, logger) {
			return fmt.Errorf("account %s does not have Name permission", tx.Input.Address)
		}
		// pubKey should be present in either "inAcc" or "tx.Input"
		if err := checkInputPubKey(inAcc, tx.Input); err != nil {
			logger.InfoMsg("Cannot find public key for input account",
				"tx_input", tx.Input)
			return err
		}
		signBytes := acm.SignBytes(exe.chainID, tx)
		amount, err := validateInput(inAcc, signBytes, tx.Input)
		if err != nil {
			logger.InfoMsg("validateInput failed",
				"tx_input", tx.Input, structure.ErrorKey, err)
			return err
		}
		if amount < tx.Fee {
			logger.InfoMsg("Sender did not send enough to cover the fee",
				"tx_input", tx.Input)
			return txs.ErrTxInsufficientFunds
		}

		// validate the input strings
		if err := tx.ValidateStrings(); err != nil {
			return err
		}

		value := amount - tx.Fee

		// let's say cost of a name for one block is len(data) + 32
		costPerBlock := txs.NameCostPerBlock(txs.NameBaseCost(tx.Name, tx.Data))
		expiresIn := value / uint64(costPerBlock)
		lastBlockHeight := exe.tip.LastBlockHeight()

		logger.TraceMsg("New NameTx",
			"value", value,
			"cost_per_block", costPerBlock,
			"expires_in", expiresIn,
			"last_block_height", lastBlockHeight)

		// check if the name exists
		entry, err := exe.nameRegCache.GetNameRegEntry(tx.Name)
		if err != nil {
			return err
		}

		if entry != nil {
			var expired bool

			// if the entry already exists, and hasn't expired, we must be owner
			if entry.Expires > lastBlockHeight {
				// ensure we are owner
				if entry.Owner != tx.Input.Address {
					return fmt.Errorf("permission denied: sender %s is trying to update a name (%s) for "+
						"which they are not an owner", tx.Input.Address, tx.Name)
				}
			} else {
				expired = true
			}

			// no value and empty data means delete the entry
			if value == 0 && len(tx.Data) == 0 {
				// maybe we reward you for telling us we can delete this crap
				// (owners if not expired, anyone if expired)
				logger.TraceMsg("Removing NameReg entry (no value and empty data in tx requests this)",
					"name", entry.Name)
				err := exe.nameRegCache.RemoveNameRegEntry(entry.Name)
				if err != nil {
					return err
				}
			} else {
				// update the entry by bumping the expiry
				// and changing the data
				if expired {
					if expiresIn < txs.MinNameRegistrationPeriod {
						return fmt.Errorf("Names must be registered for at least %d blocks", txs.MinNameRegistrationPeriod)
					}
					entry.Expires = lastBlockHeight + expiresIn
					entry.Owner = tx.Input.Address
					logger.TraceMsg("An old NameReg entry has expired and been reclaimed",
						"name", entry.Name,
						"expires_in", expiresIn,
						"owner", entry.Owner)
				} else {
					// since the size of the data may have changed
					// we use the total amount of "credit"
					oldCredit := (entry.Expires - lastBlockHeight) * txs.NameBaseCost(entry.Name, entry.Data)
					credit := oldCredit + value
					expiresIn = uint64(credit / costPerBlock)
					if expiresIn < txs.MinNameRegistrationPeriod {
						return fmt.Errorf("names must be registered for at least %d blocks", txs.MinNameRegistrationPeriod)
					}
					entry.Expires = lastBlockHeight + expiresIn
					logger.TraceMsg("Updated NameReg entry",
						"name", entry.Name,
						"expires_in", expiresIn,
						"old_credit", oldCredit,
						"value", value,
						"credit", credit)
				}
				entry.Data = tx.Data
				err := exe.nameRegCache.UpdateNameRegEntry(entry)
				if err != nil {
					return err
				}
			}
		} else {
			if expiresIn < txs.MinNameRegistrationPeriod {
				return fmt.Errorf("Names must be registered for at least %d blocks", txs.MinNameRegistrationPeriod)
			}
			// entry does not exist, so create it
			entry = &NameRegEntry{
				Name:    tx.Name,
				Owner:   tx.Input.Address,
				Data:    tx.Data,
				Expires: lastBlockHeight + expiresIn,
			}
			logger.TraceMsg("Creating NameReg entry",
				"name", entry.Name,
				"expires_in", expiresIn)
			err := exe.nameRegCache.UpdateNameRegEntry(entry)
			if err != nil {
				return err
			}
		}

		// TODO: something with the value sent?

		// Good!
		logger.TraceMsg("Incrementing sequence number for NameTx",
			"tag", "sequence",
			"account", inAcc.Address(),
			"old_sequence", inAcc.Sequence(),
			"new_sequence", inAcc.Sequence()+1)
		inAcc.IncSequence()
		err = inAcc.SubtractFromBalance(value)
		if err != nil {
			return err
		}
		exe.stateCache.UpdateAccount(inAcc)

		// TODO: maybe we want to take funds on error and allow txs in that don't do anythingi?

		if exe.eventCache != nil {
			txHash := tx.Hash(exe.chainID)
			events.PublishAccountInput(exe.eventCache, tx.Input.Address, txHash, tx, nil, "")
			events.PublishNameReg(exe.eventCache, txHash, tx)
		}

		return nil

		// Consensus related Txs inactivated for now

	case *txs.BondTx:
		from, err := getInputAccount(exe.stateCache, tx.From)
		if err != nil {
			return err
		}

		if !hasBondPermission(exe.stateCache, from, logger) {
			return fmt.Errorf("The bonder does not have permission to bond")
		}

		signBytes := acm.SignBytes(exe.chainID, tx)
		inTotal, err := validateInput(from, signBytes, tx.From)
		if err != nil {
			return err
		}

		outTotal, err := validateOutput(tx.To)
		if err != nil {
			return err
		}

		if outTotal > inTotal {
			return txs.ErrTxInsufficientFunds
		}

		fee := inTotal - outTotal
		fees += fee

		err = adjustByInput(from, tx.From, logger)
		if err != nil {
			return err
		}

		validator := exe.valPoolCache.GetMutableValidator(tx.To.Address)

		if validator == nil {
			// if bonding don't exist, sender must have CreateAccount perm
			canCreate := hasCreateAccountPermission(exe.stateCache, from, logger)

			if canCreate == false {
				return fmt.Errorf("Sender does not have permission to create accounts")
			}

			curBlockHeight := exe.tip.LastBlockHeight() + 1
			validator = acm.AsMutableValidator(acm.NewValidator(tx.BondTo, outTotal, curBlockHeight))

			err = exe.valPoolCache.AddToPool(validator)
			if err != nil {
				return err
			}
		} else {
			validator.AddStake(outTotal)

			err = exe.valPoolCache.UpdateValidator(validator)
			if err != nil {
				return err
			}
		}

		// Good! Adjust accounts
		err = exe.stateCache.UpdateAccount(from)
		if err != nil {
			return err
		}

		/// TODO: MOSTAFA
		///event....
		return nil

	case *txs.UnbondTx:
		validator := exe.valPoolCache.GetMutableValidator(tx.From.Address)
		if validator == nil {
			return txs.ErrTxInvalidAddress
		}

		to, err := state.GetMutableAccount(exe.stateCache, tx.To.Address)
		if err != nil {
			return err
		}

		// PublicKey should be present in either "account" or "input"
		if err := checkInputPubKey2(validator, tx.From); err != nil {
			return err
		}

		signBytes := acm.SignBytes(exe.chainID, tx)
		inTotal, err := validateInput2(signBytes, tx.From, validator.Sequence(), validator.Stake())
		if err != nil {
			return err
		}

		outTotal, err := validateOutput(tx.To)
		if err != nil {
			return err
		}

		if outTotal > inTotal {
			return txs.ErrTxInsufficientFunds
		}

		/// TODO:CHECK minimum stake
		if validator.MinimumStakeToUnbond() < inTotal {
			///var lastSignHeight uint64 = 0 /// TODO:MOSTAFA
			///if exe.tip.LastBlockHeight() <= lastSignHeight {
			///	return errors.New("Invalid unbond height")
			///}
		}

		fee := inTotal - outTotal
		fees += fee

		validator.IncSequence()
		validator.SubtractStake(inTotal)

		err = to.AddToBalance(outTotal)
		if err != nil {
			return err
		}

		err = exe.valPoolCache.UpdateValidator(validator)
		if err != nil {
			return err
		}

		// Good! Adjust accounts
		err = exe.stateCache.UpdateAccount(to)
		if err != nil {
			return err
		}

		/// TODO:MOSTAFA
		// Good!
		/////if exe.eventCache != nil {
		/////	exe.eventCache.Fire(txs.EventStringUnbond(), txs.EventDataTx{tx, nil, ""})
		/////}
		return nil

	case *txs.PermissionsTx:
		// Validate input
		inAcc, err := state.GetMutableAccount(exe.stateCache, tx.Input.Address)
		if err != nil {
			return err
		}
		if inAcc == nil {
			logger.InfoMsg("Cannot find input account",
				"tx_input", tx.Input)
			return txs.ErrTxInvalidAddress
		}

		err = tx.PermArgs.EnsureValid()
		if err != nil {
			return fmt.Errorf("PermissionsTx received containing invalid PermArgs: %v", err)
		}

		permFlag := tx.PermArgs.PermFlag
		// check permission
		if !HasPermission(exe.stateCache, inAcc, permFlag, logger) {
			return fmt.Errorf("account %s does not have moderator permission %s (%b)", tx.Input.Address,
				permission.PermFlagToString(permFlag), permFlag)
		}

		// pubKey should be present in either "inAcc" or "tx.Input"
		if err := checkInputPubKey(inAcc, tx.Input); err != nil {
			logger.InfoMsg("Cannot find public key for input account",
				"tx_input", tx.Input)
			return err
		}
		signBytes := acm.SignBytes(exe.chainID, tx)
		amount, err := validateInput(inAcc, signBytes, tx.Input)
		if err != nil {
			logger.InfoMsg("validateInput failed",
				"tx_input", tx.Input,
				structure.ErrorKey, err)
			return err
		}

		value := amount

		logger.TraceMsg("New PermissionsTx",
			"perm_args", tx.PermArgs.String())

		var permAcc acm.Account
		switch tx.PermArgs.PermFlag {
		case permission.HasBase:
			// this one doesn't make sense from txs
			return fmt.Errorf("HasBase is for contracts, not humans. Just look at the blockchain")
		case permission.SetBase:
			permAcc, err = mutatePermissions(exe.stateCache, *tx.PermArgs.Address,
				func(perms *ptypes.AccountPermissions) error {
					return perms.Base.Set(*tx.PermArgs.Permission, *tx.PermArgs.Value)
				})
		case permission.UnsetBase:
			permAcc, err = mutatePermissions(exe.stateCache, *tx.PermArgs.Address,
				func(perms *ptypes.AccountPermissions) error {
					return perms.Base.Unset(*tx.PermArgs.Permission)
				})
		case permission.SetGlobal:
			permAcc, err = mutatePermissions(exe.stateCache, acm.GlobalPermissionsAddress,
				func(perms *ptypes.AccountPermissions) error {
					return perms.Base.Set(*tx.PermArgs.Permission, *tx.PermArgs.Value)
				})
		case permission.HasRole:
			return fmt.Errorf("HasRole is for contracts, not humans. Just look at the blockchain")
		case permission.AddRole:
			permAcc, err = mutatePermissions(exe.stateCache, *tx.PermArgs.Address,
				func(perms *ptypes.AccountPermissions) error {
					if !perms.AddRole(*tx.PermArgs.Role) {
						return fmt.Errorf("role (%s) already exists for account %s",
							*tx.PermArgs.Role, *tx.PermArgs.Address)
					}
					return nil
				})
		case permission.RemoveRole:
			permAcc, err = mutatePermissions(exe.stateCache, *tx.PermArgs.Address,
				func(perms *ptypes.AccountPermissions) error {
					if !perms.RmRole(*tx.PermArgs.Role) {
						return fmt.Errorf("role (%s) does not exist for account %s",
							*tx.PermArgs.Role, *tx.PermArgs.Address)
					}
					return nil
				})
		default:
			return fmt.Errorf("invalid permission function: %s", permission.PermFlagToString(permFlag))
		}

		// TODO: maybe we want to take funds on error and allow txs in that don't do anythingi?
		if err != nil {
			return err
		}

		// Good!
		logger.TraceMsg("Incrementing sequence number for PermissionsTx",
			"tag", "sequence",
			"account", inAcc.Address(),
			"old_sequence", inAcc.Sequence(),
			"new_sequence", inAcc.Sequence()+1)
		inAcc.IncSequence()
		err = inAcc.SubtractFromBalance(value)
		if err != nil {
			return err
		}
		exe.stateCache.UpdateAccount(inAcc)
		if permAcc != nil {
			exe.stateCache.UpdateAccount(permAcc)
		}

		if exe.eventCache != nil {
			txHash := tx.Hash(exe.chainID)
			events.PublishAccountInput(exe.eventCache, tx.Input.Address, txHash, tx, nil, "")
			events.PublishPermissions(exe.eventCache, permission.PermFlagToString(permFlag), txHash, tx)
		}

		return nil

	case *txs.SortitionTx:
		const sortitionThreshold uint64 = 10

		/// Check if sortition tx belongs to next height, otherwise disard it
		curBlockHeight := exe.tip.LastBlockHeight() + 1
		if tx.Height < curBlockHeight-sortitionThreshold {
			return errors.New("Invalid block height")
		}

		validator := exe.valPoolCache.GetMutableValidator(tx.PublicKey.Address())
		if validator == nil {
			return errors.New("This address is not a validator")
		}

		/// validators should not submit sortition before first 10 round after bonding
		if tx.Height < validator.BondingHeight()+sortitionThreshold {
			return errors.New("Invalid block height")
		}

		isInSet := exe.tip.ValidatorSet().IsValidatorInSet(tx.PublicKey.Address())
		if isInSet == true {
			return errors.New("This validator is already in set")
		}

		/// Check if the tx is signed by transmitter and it is valid
		signBytes := acm.SignBytes(exe.chainID, tx)
		if !tx.PublicKey.VerifyBytes(signBytes, tx.Signature) {
			return txs.ErrTxInvalidSignature
		}

		/// TODO:Check bonding_height to prevent attack

		/// Verify the sortition
		var prevBlockHeight = int64(tx.Height) - 1
		result, err := rpc_core.Block(&prevBlockHeight)
		if err != nil {
			return errors.New("Invalid block height")
		}

		prevBlockHash := result.Block.Hash()
		isValid := exe.tip.VerifySortition(prevBlockHash, tx.PublicKey, tx.Index, tx.Proof)
		if isValid == false {
			return errors.New("Sortition transaction is invalid")
		}

		validator.IncSequence()

		exe.valPoolCache.AddToSet(validator)

		return nil

	default:
		// binary decoding should not let this happen
		return fmt.Errorf("unknown Tx type: %#v", tx)
	}
}

func mutatePermissions(stateReader state.Reader, address acm.Address,
	mutator func(*ptypes.AccountPermissions) error) (acm.Account, error) {

	account, err := stateReader.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("could not get account at address %s in order to alter permissions", address)
	}
	mutableAccount := acm.AsMutableAccount(account)

	return mutableAccount, mutator(mutableAccount.MutablePermissions())
}

// The accounts from the TxInputs must either already have
// acm.PublicKey().(type) != nil, (it must be known),
// or it must be specified in the TxInput.  If redeclared,
// the TxInput is modified and input.PublicKey set to nil.
func getInputAccounts(accountGetter state.AccountGetter,
	inputs []txs.TxInput) (map[acm.Address]acm.MutableAccount, error) {

	accounts := map[acm.Address]acm.MutableAccount{}
	for _, input := range inputs {
		// Account shouldn't be duplicated
		if _, ok := accounts[input.Address]; ok {
			return nil, txs.ErrTxDuplicateAddress
		}

		account, err := getInputAccount(accountGetter, input)
		if err != nil {
			return nil, err
		}
		accounts[input.Address] = account
	}
	return accounts, nil
}

func getInputAccount(accountGetter state.AccountGetter,
	input txs.TxInput) (acm.MutableAccount, error) {

	account, err := state.GetMutableAccount(accountGetter, input.Address)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, txs.ErrTxInvalidAddress
	}
	// PublicKey should be present in either "account" or "input"
	if err := checkInputPubKey(account, input); err != nil {
		return nil, err
	}

	return account, nil
}

func getOrMakeOutputs(accountGetter state.AccountGetter, accs map[acm.Address]acm.MutableAccount,
	outs []txs.TxOutput, logger *logging.Logger) (map[acm.Address]acm.MutableAccount, error) {
	if accs == nil {
		accs = make(map[acm.Address]acm.MutableAccount)
	}

	// we should err if an account is being created but the inputs don't have permission
	var checkedCreatePerms bool
	for _, out := range outs {
		// Account shouldn't be duplicated
		if _, ok := accs[out.Address]; ok {
			return nil, txs.ErrTxDuplicateAddress
		}
		acc, err := state.GetMutableAccount(accountGetter, out.Address)
		if err != nil {
			return nil, err
		}
		// output account may be nil (new)
		if acc == nil {
			if !checkedCreatePerms {
				if !haveCreateAccountPermission(accountGetter, accs, logger) {
					return nil, fmt.Errorf("at least one input does not have permission to create accounts")
				}
				checkedCreatePerms = true
			}
			acc = acm.ConcreteAccount{
				Address:     out.Address,
				Sequence:    0,
				Balance:     0,
				Permissions: permission.ZeroAccountPermissions,
			}.MutableAccount()
		}
		accs[out.Address] = acc
	}
	return accs, nil
}

// Since all ethereum accounts implicitly exist we sometimes lazily create an Account object to represent them
// only when needed. Sometimes we need to create an unknown Account knowing only its address (which is expected to
// be a deterministic hash of its associated public key) and not its public key. When we eventually receive a
// transaction acting on behalf of that account we will be given a public key that we can check matches the address.
// If it does then we will associate the public key with the stub account already registered in the system once and
// for all time.
func checkInputPubKey(account acm.MutableAccount, input txs.TxInput) error {
	if account.PublicKey().Unwrap() == nil {
		if input.PublicKey.Unwrap() == nil {
			return txs.ErrTxUnknownPubKey
		}
		addressFromPubKey := input.PublicKey.Address()
		addressFromAccount := account.Address()
		if addressFromPubKey != addressFromAccount {
			return txs.ErrTxInvalidPubKey
		}
		account.SetPublicKey(input.PublicKey)
	} else {
		input.PublicKey = acm.PublicKey{}
	}
	return nil
}

/// TODO: MOSTAFA => remove above method
func checkInputPubKey2(account acm.Addressable, input txs.TxInput) error {
	if input.PublicKey.Unwrap() == nil {
		return txs.ErrTxUnknownPubKey
	}
	addressFromPubKey := input.PublicKey.Address()
	addressFromAccount := account.Address()
	if addressFromPubKey != addressFromAccount {
		return txs.ErrTxInvalidPubKey
	}
	return nil
}

func validateInputs(accounts map[acm.Address]acm.MutableAccount, signBytes []byte, inputs []txs.TxInput) (uint64, error) {
	total := uint64(0)
	for _, input := range inputs {
		account := accounts[input.Address]
		if account == nil {
			return 0, fmt.Errorf("validateInputs() expects account in accounts, but account %s not found", input.Address)
		}
		amount, err := validateInput(account, signBytes, input)
		if err != nil {
			return 0, err
		}
		// Good. Add amount to total
		total += amount
	}
	return total, nil
}

func validateInput(account acm.MutableAccount, signBytes []byte, input txs.TxInput) (uint64, error) {
	// Check TxInput basic
	if err := input.ValidateBasic(); err != nil {
		return 0, err
	}
	// Check signatures
	if !account.PublicKey().VerifyBytes(signBytes, input.Signature) {
		return 0, txs.ErrTxInvalidSignature
	}
	// Check sequences
	if account.Sequence()+1 != uint64(input.Sequence) {
		return 0, txs.ErrTxInvalidSequence{
			Got:      input.Sequence,
			Expected: input.Sequence + uint64(1),
		}
	}
	// Check amount
	if account.Balance() < uint64(input.Amount) {
		return 0, txs.ErrTxInsufficientFunds
	}
	return input.Amount, nil
}

/// TODO: MOSTAFA => remove above method
func validateInput2(signBytes []byte, input txs.TxInput, sequence, balance uint64) (uint64, error) {
	// Check TxInput basic
	if err := input.ValidateBasic(); err != nil {
		return 0, err
	}
	// Check signatures
	if !input.PublicKey.VerifyBytes(signBytes, input.Signature) {
		return 0, txs.ErrTxInvalidSignature
	}
	// Check sequences
	if sequence+1 != uint64(input.Sequence) {
		return 0, txs.ErrTxInvalidSequence{
			Got:      input.Sequence,
			Expected: input.Sequence + uint64(1),
		}
	}
	// Check amount
	if balance < uint64(input.Amount) {
		return 0, txs.ErrTxInsufficientFunds
	}
	return input.Amount, nil
}

func validateOutputs(outputs []txs.TxOutput) (uint64, error) {
	total := uint64(0)
	for _, output := range outputs {
		// Check TxOutput basic
		amount, err := validateOutput(output)
		if err != nil {
			return 0, err
		}
		// Good. Add amount to total
		total += amount
	}
	return total, nil
}

func validateOutput(output txs.TxOutput) (uint64, error) {
	// Check TxOutput basic
	if err := output.ValidateBasic(); err != nil {
		return 0, err
	}

	return output.Amount, nil
}

func adjustByInputs(accounts map[acm.Address]acm.MutableAccount, inputs []txs.TxInput, logger *logging.Logger) error {
	for _, input := range inputs {
		account := accounts[input.Address]
		err := adjustByInput(account, input, logger)

		if err != nil {
			return err
		}
	}
	return nil
}

func adjustByInput(account acm.MutableAccount, input txs.TxInput, logger *logging.Logger) error {
	if account.Address() != input.Address {
		return fmt.Errorf("Invalid Input data")
	}

	if account.Balance() < input.Amount {
		return fmt.Errorf("No sufficient funds for account %s, only has balance %v and "+
			"we are deducting %v", input.Address, account.Balance(), input.Amount)
	}
	err := account.SubtractFromBalance(input.Amount)
	if err != nil {
		return err
	}
	logger.TraceMsg("Incrementing sequence number for SendTx (adjustByInputs)",
		"tag", "sequence",
		"account", account.Address(),
		"old_sequence", account.Sequence(),
		"new_sequence", account.Sequence()+1)
	account.IncSequence()

	return nil
}

func adjustByOutputs(accounts map[acm.Address]acm.MutableAccount, outputs []txs.TxOutput) error {
	for _, output := range outputs {

		account := accounts[output.Address]
		if account == nil {
			return fmt.Errorf("adjustByOutputs() expects account in accounts, but account %s not found",
				output.Address)
		}

		err := adjustByOutput(account, output)
		if err != nil {
			return err
		}
	}
	return nil
}

func adjustByOutput(account acm.MutableAccount, output txs.TxOutput) error {
	if account.Address() != output.Address {
		return fmt.Errorf("Invalid Input data")
	}

	err := account.AddToBalance(output.Amount)
	if err != nil {
		return err
	}

	return nil
}

//---------------------------------------------------------------

// Get permission on an account or fall back to global value
func HasPermission(accountGetter state.AccountGetter, acc acm.Account, perm ptypes.PermFlag, logger *logging.Logger) bool {
	if perm > permission.AllPermFlags {
		logger.InfoMsg(
			fmt.Sprintf("HasPermission called on invalid permission 0b%b (invalid) > 0b%b (maximum) ",
				perm, permission.AllPermFlags),
			"invalid_permission", perm,
			"maximum_permission", permission.AllPermFlags)
		return false
	}

	permString := permission.String(perm)

	v, err := acc.Permissions().Base.Compose(state.GlobalAccountPermissions(accountGetter).Base).Get(perm)
	if err != nil {
		logger.TraceMsg("Error obtaining permission value (will default to false/deny)",
			"perm_flag", permString,
			structure.ErrorKey, err)
	}

	if v {
		logger.TraceMsg("Account has permission",
			"account_address", acc.Address,
			"perm_flag", permString)
	} else {
		logger.TraceMsg("Account does not have permission",
			"account_address", acc.Address,
			"perm_flag", permString)
	}
	return v
}

// TODO: for debug log the failed accounts
func hasSendPermission(accountGetter state.AccountGetter, accs map[acm.Address]acm.MutableAccount,
	logger *logging.Logger) bool {
	for _, acc := range accs {
		if !HasPermission(accountGetter, acc, permission.Send, logger) {
			return false
		}
	}
	return true
}

func hasNamePermission(accountGetter state.AccountGetter, acc acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.Name, logger)
}

func hasCallPermission(accountGetter state.AccountGetter, acc acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.Call, logger)
}

func hasCreateContractPermission(accountGetter state.AccountGetter, acc acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.CreateContract, logger)
}

func hasCreateAccountPermission(accountGetter state.AccountGetter, acc acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.CreateAccount, logger)
}

func haveCreateAccountPermission(accountGetter state.AccountGetter, accs map[acm.Address]acm.MutableAccount,
	logger *logging.Logger) bool {
	for _, acc := range accs {
		if !hasCreateAccountPermission(accountGetter, acc, logger) {
			return false
		}
	}
	return true
}

func hasBondPermission(accountGetter state.AccountGetter, acc acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.Bond, logger)
}

func hasBondOrSendPermission(accountGetter state.AccountGetter, accs map[acm.Address]acm.MutableAccount,
	logger *logging.Logger) bool {
	for _, acc := range accs {
		if !HasPermission(accountGetter, acc, permission.Bond, logger) {
			if !HasPermission(accountGetter, acc, permission.Send, logger) {
				return false
			}
		}
	}
	return true
}

//-----------------------------------------------------------------------------

type InvalidTxError struct {
	Tx     txs.Tx
	Reason error
}

func (txErr InvalidTxError) Error() string {
	return fmt.Sprintf("Invalid tx: [%v] reason: [%v]", txErr.Tx, txErr.Reason)
}
