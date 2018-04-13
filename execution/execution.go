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
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/hyperledger/burrow/txs"
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
	// save state to disk
	err = exe.state.Save()
	if err != nil {
		return nil, err
	}
	// flush events to listeners
	exe.eventCache.Flush()
	return exe.state.Hash(), nil
}

func (exe *executor) Reset() error {
	// As with Commit() we do not take the write lock here
	exe.stateCache.Reset(exe.state)
	exe.nameRegCache.Reset(exe.state)
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
		accounts, err := getInputs(exe.stateCache, tx.Inputs)
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
			for _, i := range tx.Inputs {
				events.PublishAccountInput(exe.eventCache, i.Address, txHash, tx, nil, "")
			}

			for _, o := range tx.Outputs {
				events.PublishAccountOutput(exe.eventCache, o.Address, txHash, tx, nil, "")
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
		err = validateInput(inAcc, signBytes, tx.Input)
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
		value := tx.Input.Amount - tx.Fee

		logger.TraceMsg("Incrementing sequence number for CallTx",
			"tag", "sequence",
			"account", inAcc.Address(),
			"old_sequence", inAcc.Sequence(),
			"new_sequence", inAcc.Sequence()+1)

		inAcc, err = inAcc.IncSequence().SubtractFromBalance(tx.Fee)
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

			// Run VM call and sync txCache to exe.blockCache.
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
			inAcc, err = inAcc.SubtractFromBalance(value)
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
		err = validateInput(inAcc, signBytes, tx.Input)
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

		// validate the input strings
		if err := tx.ValidateStrings(); err != nil {
			return err
		}

		value := tx.Input.Amount - tx.Fee

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
		inAcc, err = inAcc.SubtractFromBalance(value)
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
		// TODO!
		/*
			case *txs.BondTx:
						valInfo := exe.blockCache.State().GetValidatorInfo(tx.PublicKey().Address())
						if valInfo != nil {
							// TODO: In the future, check that the validator wasn't destroyed,
							// add funds, merge UnbondTo outputs, and unbond validator.
							return errors.New("Adding coins to existing validators not yet supported")
						}

						accounts, err := getInputs(exe.blockCache, tx.Inputs)
						if err != nil {
							return err
						}

						// add outputs to accounts map
						// if any outputs don't exist, all inputs must have CreateAccount perm
						// though outputs aren't created until unbonding/release time
						canCreate := hasCreateAccountPermission(exe.blockCache, accounts)
						for _, out := range tx.UnbondTo {
							acc := exe.blockCache.GetAccount(out.Address)
							if acc == nil && !canCreate {
								return fmt.Errorf("At least one input does not have permission to create accounts")
							}
						}

						bondAcc := exe.blockCache.GetAccount(tx.PublicKey().Address())
						if !hasBondPermission(exe.blockCache, bondAcc) {
							return fmt.Errorf("The bonder does not have permission to bond")
						}

						if !hasBondOrSendPermission(exe.blockCache, accounts) {
							return fmt.Errorf("At least one input lacks permission to bond")
						}

						signBytes := acm.SignBytes(exe.chainID, tx)
						inTotal, err := validateInputs(accounts, signBytes, tx.Inputs)
						if err != nil {
							return err
						}
						if !tx.PublicKey().VerifyBytes(signBytes, tx.Signature) {
							return txs.ErrTxInvalidSignature
						}
						outTotal, err := validateOutputs(tx.UnbondTo)
						if err != nil {
							return err
						}
						if outTotal > inTotal {
							return txs.ErrTxInsufficientFunds
						}
						fee := inTotal - outTotal
						fees += fee

						// Good! Adjust accounts
						adjustByInputs(accounts, tx.Inputs)
						for _, acc := range accounts {
							exe.blockCache.UpdateAccount(acc)
						}
						// Add ValidatorInfo
						_s.SetValidatorInfo(&txs.ValidatorInfo{
							Address:         tx.PublicKey().Address(),
							PublicKey:          tx.PublicKey(),
							UnbondTo:        tx.UnbondTo,
							FirstBondHeight: _s.lastBlockHeight + 1,
							FirstBondAmount: outTotal,
						})
						// Add Validator
						added := _s.BondedValidators.Add(&txs.Validator{
							Address:     tx.PublicKey().Address(),
							PublicKey:      tx.PublicKey(),
							BondHeight:  _s.lastBlockHeight + 1,
							VotingPower: outTotal,
							Accum:       0,
						})
						if !added {
							PanicCrisis("Failed to add validator")
						}
						if exe.eventCache != nil {
							// TODO: fire for all inputs
							exe.eventCache.Fire(txs.EventStringBond(), txs.EventDataTx{tx, nil, ""})
						}
						return nil

					case *txs.UnbondTx:
						// The validator must be active
						_, val := _s.BondedValidators.GetByAddress(tx.Address)
						if val == nil {
							return txs.ErrTxInvalidAddress
						}

						// Verify the signature
						signBytes := acm.SignBytes(exe.chainID, tx)
						if !val.PublicKey().VerifyBytes(signBytes, tx.Signature) {
							return txs.ErrTxInvalidSignature
						}

						// tx.Height must be greater than val.LastCommitHeight
						if tx.Height <= val.LastCommitHeight {
							return errors.New("Invalid unbond height")
						}

						// Good!
						_s.unbondValidator(val)
						if exe.eventCache != nil {
							exe.eventCache.Fire(txs.EventStringUnbond(), txs.EventDataTx{tx, nil, ""})
						}
						return nil

					case *txs.RebondTx:
						// The validator must be inactive
						_, val := _s.UnbondingValidators.GetByAddress(tx.Address)
						if val == nil {
							return txs.ErrTxInvalidAddress
						}

						// Verify the signature
						signBytes := acm.SignBytes(exe.chainID, tx)
						if !val.PublicKey().VerifyBytes(signBytes, tx.Signature) {
							return txs.ErrTxInvalidSignature
						}

						// tx.Height must be in a suitable range
						minRebondHeight := _s.lastBlockHeight - (validatorTimeoutBlocks / 2)
						maxRebondHeight := _s.lastBlockHeight + 2
						if !((minRebondHeight <= tx.Height) && (tx.Height <= maxRebondHeight)) {
							return errors.New(Fmt("Rebond height not in range.  Expected %v <= %v <= %v",
								minRebondHeight, tx.Height, maxRebondHeight))
						}

						// Good!
						_s.rebondValidator(val)
						if exe.eventCache != nil {
							exe.eventCache.Fire(txs.EventStringRebond(), txs.EventDataTx{tx, nil, ""})
						}
						return nil

		*/

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
		err = validateInput(inAcc, signBytes, tx.Input)
		if err != nil {
			logger.InfoMsg("validateInput failed",
				"tx_input", tx.Input,
				structure.ErrorKey, err)
			return err
		}

		value := tx.Input.Amount

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
		inAcc, err = inAcc.SubtractFromBalance(value)
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

// ExecBlock stuff is now taken care of by the consensus engine.
// But we leave here for now for reference when we have to do validator updates

/*

// NOTE: If an error occurs during block execution, state will be left
// at an invalid state.  Copy the state before calling ExecBlock!
func ExecBlock(s *State, block *txs.Block, blockPartsHeader txs.PartSetHeader) error {
	err := execBlock(s, block, blockPartsHeader)
	if err != nil {
		return err
	}
	// State.Hash should match block.StateHash
	stateHash := s.Hash()
	if !bytes.Equal(stateHash, block.StateHash) {
		return errors.New(Fmt("Invalid state hash. Expected %X, got %X",
			stateHash, block.StateHash))
	}
	return nil
}

// executes transactions of a block, does not check block.StateHash
// NOTE: If an error occurs during block execution, state will be left
// at an invalid state.  Copy the state before calling execBlock!
func execBlock(s *State, block *txs.Block, blockPartsHeader txs.PartSetHeader) error {
	// Basic block validation.
	err := block.ValidateBasic(s.chainID, s.lastBlockHeight, s.lastBlockAppHash, s.LastBlockParts, s.lastBlockTime)
	if err != nil {
		return err
	}

	// Validate block LastValidation.
	if block.Height == 1 {
		if len(block.LastValidation.Precommits) != 0 {
			return errors.New("Block at height 1 (first block) should have no LastValidation precommits")
		}
	} else {
		if len(block.LastValidation.Precommits) != s.LastBondedValidators.Size() {
			return errors.New(Fmt("Invalid block validation size. Expected %v, got %v",
				s.LastBondedValidators.Size(), len(block.LastValidation.Precommits)))
		}
		err := s.LastBondedValidators.VerifyValidation(
			s.chainID, s.lastBlockAppHash, s.LastBlockParts, block.Height-1, block.LastValidation)
		if err != nil {
			return err
		}
	}

	// Update Validator.LastCommitHeight as necessary.
	for i, precommit := range block.LastValidation.Precommits {
		if precommit == nil {
			continue
		}
		_, val := s.LastBondedValidators.GetByIndex(i)
		if val == nil {
			PanicCrisis(Fmt("Failed to fetch validator at index %v", i))
		}
		if _, val_ := s.BondedValidators.GetByAddress(val.Address); val_ != nil {
			val_.LastCommitHeight = block.Height - 1
			updated := s.BondedValidators.Update(val_)
			if !updated {
				PanicCrisis("Failed to update bonded validator LastCommitHeight")
			}
		} else if _, val_ := s.UnbondingValidators.GetByAddress(val.Address); val_ != nil {
			val_.LastCommitHeight = block.Height - 1
			updated := s.UnbondingValidators.Update(val_)
			if !updated {
				PanicCrisis("Failed to update unbonding validator LastCommitHeight")
			}
		} else {
			PanicCrisis("Could not find validator")
		}
	}

	// Remember LastBondedValidators
	s.LastBondedValidators = s.BondedValidators.Copy()

	// Create BlockCache to cache changes to state.
	blockCache := NewBlockCache(s)

	// Execute each tx
	for _, tx := range block.Data.Txs {
		err := ExecTx(blockCache, tx, true, s.eventCache)
		if err != nil {
			return InvalidTxError{tx, err}
		}
	}

	// Now sync the BlockCache to the backend.
	blockCache.Sync()

	// If any unbonding periods are over,
	// reward account with bonded coins.
	toRelease := []*txs.Validator{}
	s.UnbondingValidators.Iterate(func(index int, val *txs.Validator) bool {
		if val.UnbondHeight+unbondingPeriodBlocks < block.Height {
			toRelease = append(toRelease, val)
		}
		return false
	})
	for _, val := range toRelease {
		s.releaseValidator(val)
	}

	// If any validators haven't signed in a while,
	// unbond them, they have timed out.
	toTimeout := []*txs.Validator{}
	s.BondedValidators.Iterate(func(index int, val *txs.Validator) bool {
		lastActivityHeight := MaxInt(val.BondHeight, val.LastCommitHeight)
		if lastActivityHeight+validatorTimeoutBlocks < block.Height {
			log.Notice("Validator timeout", "validator", val, "height", block.Height)
			toTimeout = append(toTimeout, val)
		}
		return false
	})
	for _, val := range toTimeout {
		s.unbondValidator(val)
	}

	// Increment validator AccumPowers
	s.BondedValidators.IncrementAccum(1)
	s.lastBlockHeight = block.Height
	s.lastBlockAppHash = block.Hash()
	s.LastBlockParts = blockPartsHeader
	s.lastBlockTime = block.Time
	return nil
}
*/

// The accounts from the TxInputs must either already have
// acm.PublicKey().(type) != nil, (it must be known),
// or it must be specified in the TxInput.  If redeclared,
// the TxInput is modified and input.PublicKey() set to nil.
func getInputs(accountGetter state.AccountGetter,
	ins []*txs.TxInput) (map[acm.Address]acm.MutableAccount, error) {

	accounts := map[acm.Address]acm.MutableAccount{}
	for _, in := range ins {
		// Account shouldn't be duplicated
		if _, ok := accounts[in.Address]; ok {
			return nil, txs.ErrTxDuplicateAddress
		}
		acc, err := state.GetMutableAccount(accountGetter, in.Address)
		if err != nil {
			return nil, err
		}
		if acc == nil {
			return nil, txs.ErrTxInvalidAddress
		}
		// PublicKey should be present in either "account" or "in"
		if err := checkInputPubKey(acc, in); err != nil {
			return nil, err
		}
		accounts[in.Address] = acc
	}
	return accounts, nil
}

func getOrMakeOutputs(accountGetter state.AccountGetter, accs map[acm.Address]acm.MutableAccount,
	outs []*txs.TxOutput, logger *logging.Logger) (map[acm.Address]acm.MutableAccount, error) {
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
				if !hasCreateAccountPermission(accountGetter, accs, logger) {
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
func checkInputPubKey(acc acm.MutableAccount, in *txs.TxInput) error {
	if acc.PublicKey().Unwrap() == nil {
		if in.PublicKey.Unwrap() == nil {
			return txs.ErrTxUnknownPubKey
		}
		addressFromPubKey := in.PublicKey.Address()
		addressFromAccount := acc.Address()
		if addressFromPubKey != addressFromAccount {
			return txs.ErrTxInvalidPubKey
		}
		acc.SetPublicKey(in.PublicKey)
	} else {
		in.PublicKey = acm.PublicKey{}
	}
	return nil
}

func validateInputs(accs map[acm.Address]acm.MutableAccount, signBytes []byte, ins []*txs.TxInput) (uint64, error) {
	total := uint64(0)
	for _, in := range ins {
		acc := accs[in.Address]
		if acc == nil {
			return 0, fmt.Errorf("validateInputs() expects account in accounts, but account %s not found", in.Address)
		}
		err := validateInput(acc, signBytes, in)
		if err != nil {
			return 0, err
		}
		// Good. Add amount to total
		total += in.Amount
	}
	return total, nil
}

func validateInput(acc acm.MutableAccount, signBytes []byte, in *txs.TxInput) error {
	// Check TxInput basic
	if err := in.ValidateBasic(); err != nil {
		return err
	}
	// Check signatures
	if !acc.PublicKey().VerifyBytes(signBytes, in.Signature) {
		return txs.ErrTxInvalidSignature
	}
	// Check sequences
	if acc.Sequence()+1 != uint64(in.Sequence) {
		return txs.ErrTxInvalidSequence{
			Got:      in.Sequence,
			Expected: acc.Sequence() + uint64(1),
		}
	}
	// Check amount
	if acc.Balance() < uint64(in.Amount) {
		return txs.ErrTxInsufficientFunds
	}
	return nil
}

func validateOutputs(outs []*txs.TxOutput) (uint64, error) {
	total := uint64(0)
	for _, out := range outs {
		// Check TxOutput basic
		if err := out.ValidateBasic(); err != nil {
			return 0, err
		}
		// Good. Add amount to total
		total += out.Amount
	}
	return total, nil
}

func adjustByInputs(accs map[acm.Address]acm.MutableAccount, ins []*txs.TxInput, logger *logging.Logger) error {
	for _, in := range ins {
		acc := accs[in.Address]
		if acc == nil {
			return fmt.Errorf("adjustByInputs() expects account in accounts, but account %s not found", in.Address)
		}
		if acc.Balance() < in.Amount {
			panic("adjustByInputs() expects sufficient funds")
			return fmt.Errorf("adjustByInputs() expects sufficient funds but account %s only has balance %v and "+
				"we are deducting %v", in.Address, acc.Balance(), in.Amount)
		}
		acc, err := acc.SubtractFromBalance(in.Amount)
		if err != nil {
			return err
		}
		logger.TraceMsg("Incrementing sequence number for SendTx (adjustByInputs)",
			"tag", "sequence",
			"account", acc.Address(),
			"old_sequence", acc.Sequence(),
			"new_sequence", acc.Sequence()+1)
		acc.IncSequence()
	}
	return nil
}

func adjustByOutputs(accs map[acm.Address]acm.MutableAccount, outs []*txs.TxOutput) error {
	for _, out := range outs {
		acc := accs[out.Address]
		if acc == nil {
			return fmt.Errorf("adjustByOutputs() expects account in accounts, but account %s not found",
				out.Address)
		}
		_, err := acc.AddToBalance(out.Amount)
		if err != nil {
			return err
		}
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

func hasCreateAccountPermission(accountGetter state.AccountGetter, accs map[acm.Address]acm.MutableAccount,
	logger *logging.Logger) bool {
	for _, acc := range accs {
		if !HasPermission(accountGetter, acc, permission.CreateAccount, logger) {
			return false
		}
	}
	return true
}

func hasBondPermission(accountGetter state.AccountGetter, acc acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.Bond, logger)
}

func hasBondOrSendPermission(accountGetter state.AccountGetter, accs map[acm.Address]acm.Account,
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
