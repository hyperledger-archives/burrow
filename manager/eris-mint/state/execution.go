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

package state

import (
	"bytes"
	"errors"
	"fmt"

	acm "github.com/monax/burrow/account"
	"github.com/monax/burrow/common/sanity"
	core_types "github.com/monax/burrow/core/types"
	logging_types "github.com/monax/burrow/logging/types"
	"github.com/monax/burrow/manager/eris-mint/evm"
	ptypes "github.com/monax/burrow/permission/types" // for GlobalPermissionAddress ...
	"github.com/monax/burrow/txs"
	. "github.com/monax/burrow/word256"

	"github.com/tendermint/go-events"
)

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
	err := block.ValidateBasic(s.ChainID, s.LastBlockHeight, s.LastBlockHash, s.LastBlockParts, s.LastBlockTime)
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
			s.ChainID, s.LastBlockHash, s.LastBlockParts, block.Height-1, block.LastValidation)
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
		err := ExecTx(blockCache, tx, true, s.evc)
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
	s.LastBlockHeight = block.Height
	s.LastBlockHash = block.Hash()
	s.LastBlockParts = blockPartsHeader
	s.LastBlockTime = block.Time
	return nil
}
*/

// The accounts from the TxInputs must either already have
// acm.PubKey.(type) != nil, (it must be known),
// or it must be specified in the TxInput.  If redeclared,
// the TxInput is modified and input.PubKey set to nil.
func getInputs(state AccountGetter, ins []*txs.TxInput) (map[string]*acm.Account, error) {
	accounts := map[string]*acm.Account{}
	for _, in := range ins {
		// Account shouldn't be duplicated
		if _, ok := accounts[string(in.Address)]; ok {
			return nil, txs.ErrTxDuplicateAddress
		}
		acc := state.GetAccount(in.Address)
		if acc == nil {
			return nil, txs.ErrTxInvalidAddress
		}
		// PubKey should be present in either "account" or "in"
		if err := checkInputPubKey(acc, in); err != nil {
			return nil, err
		}
		accounts[string(in.Address)] = acc
	}
	return accounts, nil
}

func getOrMakeOutputs(state AccountGetter, accounts map[string]*acm.Account, outs []*txs.TxOutput) (map[string]*acm.Account, error) {
	if accounts == nil {
		accounts = make(map[string]*acm.Account)
	}

	// we should err if an account is being created but the inputs don't have permission
	var checkedCreatePerms bool
	for _, out := range outs {
		// Account shouldn't be duplicated
		if _, ok := accounts[string(out.Address)]; ok {
			return nil, txs.ErrTxDuplicateAddress
		}
		acc := state.GetAccount(out.Address)
		// output account may be nil (new)
		if acc == nil {
			if !checkedCreatePerms {
				if !hasCreateAccountPermission(state, accounts) {
					return nil, fmt.Errorf("At least one input does not have permission to create accounts")
				}
				checkedCreatePerms = true
			}
			acc = &acm.Account{
				Address:     out.Address,
				PubKey:      nil,
				Sequence:    0,
				Balance:     0,
				Permissions: ptypes.ZeroAccountPermissions,
			}
		}
		accounts[string(out.Address)] = acc
	}
	return accounts, nil
}

// Since all ethereum accounts implicitly exist we sometimes lazily create an Account object to represent them
// only when needed. Sometimes we need to create an unknown Account knowing only its address (which is expected to
// be a deterministic hash of its associated public key) and not its public key. When we eventually receive a
// transaction acting on behalf of that account we will be given a public key that we can check matches the address.
// If it does then we will associate the public key with the stub account already registered in the system once and
// for all time.
func checkInputPubKey(acc *acm.Account, in *txs.TxInput) error {
	if acc.PubKey == nil {
		if in.PubKey == nil {
			return txs.ErrTxUnknownPubKey
		}
		if !bytes.Equal(in.PubKey.Address(), acc.Address) {
			return txs.ErrTxInvalidPubKey
		}
		acc.PubKey = in.PubKey
	} else {
		in.PubKey = nil
	}
	return nil
}

func validateInputs(accounts map[string]*acm.Account, signBytes []byte, ins []*txs.TxInput) (total int64, err error) {
	for _, in := range ins {
		acc := accounts[string(in.Address)]
		if acc == nil {
			sanity.PanicSanity("validateInputs() expects account in accounts")
		}
		err = validateInput(acc, signBytes, in)
		if err != nil {
			return
		}
		// Good. Add amount to total
		total += in.Amount
	}
	return total, nil
}

func validateInput(acc *acm.Account, signBytes []byte, in *txs.TxInput) (err error) {
	// Check TxInput basic
	if err := in.ValidateBasic(); err != nil {
		return err
	}
	// Check signatures
	if !acc.PubKey.VerifyBytes(signBytes, in.Signature) {
		return txs.ErrTxInvalidSignature
	}
	// Check sequences
	if acc.Sequence+1 != in.Sequence {
		return txs.ErrTxInvalidSequence{
			Got:      in.Sequence,
			Expected: acc.Sequence + 1,
		}
	}
	// Check amount
	if acc.Balance < in.Amount {
		return txs.ErrTxInsufficientFunds
	}
	return nil
}

func validateOutputs(outs []*txs.TxOutput) (total int64, err error) {
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

func adjustByInputs(accounts map[string]*acm.Account, ins []*txs.TxInput) {
	for _, in := range ins {
		acc := accounts[string(in.Address)]
		if acc == nil {
			sanity.PanicSanity("adjustByInputs() expects account in accounts")
		}
		if acc.Balance < in.Amount {
			sanity.PanicSanity("adjustByInputs() expects sufficient funds")
		}
		acc.Balance -= in.Amount

		acc.Sequence += 1
	}
}

func adjustByOutputs(accounts map[string]*acm.Account, outs []*txs.TxOutput) {
	for _, out := range outs {
		acc := accounts[string(out.Address)]
		if acc == nil {
			sanity.PanicSanity("adjustByOutputs() expects account in accounts")
		}
		acc.Balance += out.Amount
	}
}

// If the tx is invalid, an error will be returned.
// Unlike ExecBlock(), state will not be altered.
func ExecTx(blockCache *BlockCache, tx txs.Tx, runCall bool, evc events.Fireable,
		logger logging_types.InfoTraceLogger) (err error) {

	// TODO: do something with fees
	fees := int64(0)
	_s := blockCache.State() // hack to access validators and block height

	// Exec tx
	switch tx := tx.(type) {
	case *txs.SendTx:
		accounts, err := getInputs(blockCache, tx.Inputs)
		if err != nil {
			return err
		}

		// ensure all inputs have send permissions
		if !hasSendPermission(blockCache, accounts) {
			return fmt.Errorf("At least one input lacks permission for SendTx")
		}

		// add outputs to accounts map
		// if any outputs don't exist, all inputs must have CreateAccount perm
		accounts, err = getOrMakeOutputs(blockCache, accounts, tx.Outputs)
		if err != nil {
			return err
		}

		signBytes := acm.SignBytes(_s.ChainID, tx)
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
		adjustByInputs(accounts, tx.Inputs)
		adjustByOutputs(accounts, tx.Outputs)
		for _, acc := range accounts {
			blockCache.UpdateAccount(acc)
		}

		// if the evc is nil, nothing will happen
		if evc != nil {
			for _, i := range tx.Inputs {
				evc.FireEvent(txs.EventStringAccInput(i.Address), txs.EventDataTx{tx, nil, ""})
			}

			for _, o := range tx.Outputs {
				evc.FireEvent(txs.EventStringAccOutput(o.Address), txs.EventDataTx{tx, nil, ""})
			}
		}
		return nil

	case *txs.CallTx:
		var inAcc, outAcc *acm.Account

		// Validate input
		inAcc = blockCache.GetAccount(tx.Input.Address)
		if inAcc == nil {
			log.Info(fmt.Sprintf("Can't find in account %X", tx.Input.Address))
			return txs.ErrTxInvalidAddress
		}

		createContract := len(tx.Address) == 0
		if createContract {
			if !hasCreateContractPermission(blockCache, inAcc) {
				return fmt.Errorf("Account %X does not have CreateContract permission", tx.Input.Address)
			}
		} else {
			if !hasCallPermission(blockCache, inAcc) {
				return fmt.Errorf("Account %X does not have Call permission", tx.Input.Address)
			}
		}

		// pubKey should be present in either "inAcc" or "tx.Input"
		if err := checkInputPubKey(inAcc, tx.Input); err != nil {
			log.Info(fmt.Sprintf("Can't find pubkey for %X", tx.Input.Address))
			return err
		}
		signBytes := acm.SignBytes(_s.ChainID, tx)
		err := validateInput(inAcc, signBytes, tx.Input)
		if err != nil {
			log.Info(fmt.Sprintf("validateInput failed on %X: %v", tx.Input.Address, err))
			return err
		}
		if tx.Input.Amount < tx.Fee {
			log.Info(fmt.Sprintf("Sender did not send enough to cover the fee %X", tx.Input.Address))
			return txs.ErrTxInsufficientFunds
		}

		if !createContract {
			// Validate output
			if len(tx.Address) != 20 {
				log.Info(fmt.Sprintf("Destination address is not 20 bytes %X", tx.Address))
				return txs.ErrTxInvalidAddress
			}
			// check if its a native contract
			if vm.RegisteredNativeContract(LeftPadWord256(tx.Address)) {
				return fmt.Errorf("NativeContracts can not be called using CallTx. Use a contract or the appropriate tx type (eg. PermissionsTx, NameTx)")
			}

			// Output account may be nil if we are still in mempool and contract was created in same block as this tx
			// but that's fine, because the account will be created properly when the create tx runs in the block
			// and then this won't return nil. otherwise, we take their fee
			outAcc = blockCache.GetAccount(tx.Address)
		}

		log.Info(fmt.Sprintf("Out account: %v", outAcc))

		// Good!
		value := tx.Input.Amount - tx.Fee

		inAcc.Sequence += 1
		inAcc.Balance -= tx.Fee
		blockCache.UpdateAccount(inAcc)

		// The logic in runCall MUST NOT return.
		if runCall {

			// VM call variables
			var (
				gas     int64       = tx.GasLimit
				err     error       = nil
				caller  *vm.Account = toVMAccount(inAcc)
				callee  *vm.Account = nil // initialized below
				code    []byte      = nil
				ret     []byte      = nil
				txCache             = NewTxCache(blockCache)
				params              = vm.Params{
					BlockHeight: int64(_s.LastBlockHeight),
					BlockHash:   LeftPadWord256(_s.LastBlockHash),
					BlockTime:   _s.LastBlockTime.Unix(),
					GasLimit:    _s.GetGasLimit(),
				}
			)

			if !createContract && (outAcc == nil || len(outAcc.Code) == 0) {
				// if you call an account that doesn't exist
				// or an account with no code then we take fees (sorry pal)
				// NOTE: it's fine to create a contract and call it within one
				// block (nonce will prevent re-ordering of those txs)
				// but to create with one contract and call with another
				// you have to wait a block to avoid a re-ordering attack
				// that will take your fees
				if outAcc == nil {
					log.Info(fmt.Sprintf("%X tries to call %X but it does not exist.",
						inAcc.Address, tx.Address))
				} else {
					log.Info(fmt.Sprintf("%X tries to call %X but code is blank.",
						inAcc.Address, tx.Address))
				}
				err = txs.ErrTxInvalidAddress
				goto CALL_COMPLETE
			}

			// get or create callee
			if createContract {
				// We already checked for permission
				callee = txCache.CreateAccount(caller)
				log.Info(fmt.Sprintf("Created new contract %X", callee.Address))
				code = tx.Data
			} else {
				callee = toVMAccount(outAcc)
				log.Info(fmt.Sprintf("Calling contract %X with code %X", callee.Address, callee.Code))
				code = callee.Code
			}
			log.Info(fmt.Sprintf("Code for this contract: %X", code))

			// Run VM call and sync txCache to blockCache.
			{ // Capture scope for goto.
				// Write caller/callee to txCache.
				txCache.UpdateAccount(caller)
				txCache.UpdateAccount(callee)
				vmach := vm.NewVM(txCache, params, caller.Address, txs.TxHash(_s.ChainID, tx))
				vmach.SetFireable(evc)
				// NOTE: Call() transfers the value from caller to callee iff call succeeds.
				ret, err = vmach.Call(caller, callee, code, tx.Data, value, &gas)
				if err != nil {
					// Failure. Charge the gas fee. The 'value' was otherwise not transferred.
					log.Info(fmt.Sprintf("Error on execution: %v", err))
					goto CALL_COMPLETE
				}

				log.Info("Successful execution")
				if createContract {
					callee.Code = ret
				}
				txCache.Sync()
			}

		CALL_COMPLETE: // err may or may not be nil.

		// Create a receipt from the ret and whether errored.
			log.Notice("VM call complete", "caller", caller, "callee", callee, "return", ret, "err", err)

			// Fire Events for sender and receiver
			// a separate event will be fired from vm for each additional call
			if evc != nil {
				exception := ""
				if err != nil {
					exception = err.Error()
				}
				evc.FireEvent(txs.EventStringAccInput(tx.Input.Address), txs.EventDataTx{tx, ret, exception})
				evc.FireEvent(txs.EventStringAccOutput(tx.Address), txs.EventDataTx{tx, ret, exception})
			}
		} else {
			// The mempool does not call txs until
			// the proposer determines the order of txs.
			// So mempool will skip the actual .Call(),
			// and only deduct from the caller's balance.
			inAcc.Balance -= value
			if createContract {
				inAcc.Sequence += 1 // XXX ?!
			}
			blockCache.UpdateAccount(inAcc)
		}

		return nil

	case *txs.NameTx:
		var inAcc *acm.Account

		// Validate input
		inAcc = blockCache.GetAccount(tx.Input.Address)
		if inAcc == nil {
			log.Info(fmt.Sprintf("Can't find in account %X", tx.Input.Address))
			return txs.ErrTxInvalidAddress
		}
		// check permission
		if !hasNamePermission(blockCache, inAcc) {
			return fmt.Errorf("Account %X does not have Name permission", tx.Input.Address)
		}
		// pubKey should be present in either "inAcc" or "tx.Input"
		if err := checkInputPubKey(inAcc, tx.Input); err != nil {
			log.Info(fmt.Sprintf("Can't find pubkey for %X", tx.Input.Address))
			return err
		}
		signBytes := acm.SignBytes(_s.ChainID, tx)
		err := validateInput(inAcc, signBytes, tx.Input)
		if err != nil {
			log.Info(fmt.Sprintf("validateInput failed on %X: %v", tx.Input.Address, err))
			return err
		}
		// fee is in addition to the amount which is used to determine the TTL
		if tx.Input.Amount < tx.Fee {
			log.Info(fmt.Sprintf("Sender did not send enough to cover the fee %X", tx.Input.Address))
			return txs.ErrTxInsufficientFunds
		}

		// validate the input strings
		if err := tx.ValidateStrings(); err != nil {
			return err
		}

		value := tx.Input.Amount - tx.Fee

		// let's say cost of a name for one block is len(data) + 32
		costPerBlock := txs.NameCostPerBlock(txs.NameBaseCost(tx.Name, tx.Data))
		expiresIn := int(value / costPerBlock)
		lastBlockHeight := _s.LastBlockHeight

		log.Info("New NameTx", "value", value, "costPerBlock", costPerBlock, "expiresIn", expiresIn, "lastBlock", lastBlockHeight)

		// check if the name exists
		entry := blockCache.GetNameRegEntry(tx.Name)

		if entry != nil {
			var expired bool

			// if the entry already exists, and hasn't expired, we must be owner
			if entry.Expires > lastBlockHeight {
				// ensure we are owner
				if bytes.Compare(entry.Owner, tx.Input.Address) != 0 {
					log.Info(fmt.Sprintf("Sender %X is trying to update a name (%s) for which he is not owner", tx.Input.Address, tx.Name))
					return txs.ErrTxPermissionDenied
				}
			} else {
				expired = true
			}

			// no value and empty data means delete the entry
			if value == 0 && len(tx.Data) == 0 {
				// maybe we reward you for telling us we can delete this crap
				// (owners if not expired, anyone if expired)
				log.Info("Removing namereg entry", "name", entry.Name)
				blockCache.RemoveNameRegEntry(entry.Name)
			} else {
				// update the entry by bumping the expiry
				// and changing the data
				if expired {
					if expiresIn < txs.MinNameRegistrationPeriod {
						return errors.New(fmt.Sprintf("Names must be registered for at least %d blocks", txs.MinNameRegistrationPeriod))
					}
					entry.Expires = lastBlockHeight + expiresIn
					entry.Owner = tx.Input.Address
					log.Info("An old namereg entry has expired and been reclaimed", "name", entry.Name, "expiresIn", expiresIn, "owner", entry.Owner)
				} else {
					// since the size of the data may have changed
					// we use the total amount of "credit"
					oldCredit := int64(entry.Expires-lastBlockHeight) * txs.NameBaseCost(entry.Name, entry.Data)
					credit := oldCredit + value
					expiresIn = int(credit / costPerBlock)
					if expiresIn < txs.MinNameRegistrationPeriod {
						return errors.New(fmt.Sprintf("Names must be registered for at least %d blocks", txs.MinNameRegistrationPeriod))
					}
					entry.Expires = lastBlockHeight + expiresIn
					log.Info("Updated namereg entry", "name", entry.Name, "expiresIn", expiresIn, "oldCredit", oldCredit, "value", value, "credit", credit)
				}
				entry.Data = tx.Data
				blockCache.UpdateNameRegEntry(entry)
			}
		} else {
			if expiresIn < txs.MinNameRegistrationPeriod {
				return errors.New(fmt.Sprintf("Names must be registered for at least %d blocks", txs.MinNameRegistrationPeriod))
			}
			// entry does not exist, so create it
			entry = &core_types.NameRegEntry{
				Name:    tx.Name,
				Owner:   tx.Input.Address,
				Data:    tx.Data,
				Expires: lastBlockHeight + expiresIn,
			}
			log.Info("Creating namereg entry", "name", entry.Name, "expiresIn", expiresIn)
			blockCache.UpdateNameRegEntry(entry)
		}

		// TODO: something with the value sent?

		// Good!
		inAcc.Sequence += 1
		inAcc.Balance -= value
		blockCache.UpdateAccount(inAcc)

		// TODO: maybe we want to take funds on error and allow txs in that don't do anythingi?

		if evc != nil {
			evc.FireEvent(txs.EventStringAccInput(tx.Input.Address), txs.EventDataTx{tx, nil, ""})
			evc.FireEvent(txs.EventStringNameReg(tx.Name), txs.EventDataTx{tx, nil, ""})
		}

		return nil

	// Consensus related Txs inactivated for now
	// TODO!
	/*
		case *txs.BondTx:
					valInfo := blockCache.State().GetValidatorInfo(tx.PubKey.Address())
					if valInfo != nil {
						// TODO: In the future, check that the validator wasn't destroyed,
						// add funds, merge UnbondTo outputs, and unbond validator.
						return errors.New("Adding coins to existing validators not yet supported")
					}

					accounts, err := getInputs(blockCache, tx.Inputs)
					if err != nil {
						return err
					}

					// add outputs to accounts map
					// if any outputs don't exist, all inputs must have CreateAccount perm
					// though outputs aren't created until unbonding/release time
					canCreate := hasCreateAccountPermission(blockCache, accounts)
					for _, out := range tx.UnbondTo {
						acc := blockCache.GetAccount(out.Address)
						if acc == nil && !canCreate {
							return fmt.Errorf("At least one input does not have permission to create accounts")
						}
					}

					bondAcc := blockCache.GetAccount(tx.PubKey.Address())
					if !hasBondPermission(blockCache, bondAcc) {
						return fmt.Errorf("The bonder does not have permission to bond")
					}

					if !hasBondOrSendPermission(blockCache, accounts) {
						return fmt.Errorf("At least one input lacks permission to bond")
					}

					signBytes := acm.SignBytes(_s.ChainID, tx)
					inTotal, err := validateInputs(accounts, signBytes, tx.Inputs)
					if err != nil {
						return err
					}
					if !tx.PubKey.VerifyBytes(signBytes, tx.Signature) {
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
						blockCache.UpdateAccount(acc)
					}
					// Add ValidatorInfo
					_s.SetValidatorInfo(&txs.ValidatorInfo{
						Address:         tx.PubKey.Address(),
						PubKey:          tx.PubKey,
						UnbondTo:        tx.UnbondTo,
						FirstBondHeight: _s.LastBlockHeight + 1,
						FirstBondAmount: outTotal,
					})
					// Add Validator
					added := _s.BondedValidators.Add(&txs.Validator{
						Address:     tx.PubKey.Address(),
						PubKey:      tx.PubKey,
						BondHeight:  _s.LastBlockHeight + 1,
						VotingPower: outTotal,
						Accum:       0,
					})
					if !added {
						PanicCrisis("Failed to add validator")
					}
					if evc != nil {
						// TODO: fire for all inputs
						evc.FireEvent(txs.EventStringBond(), txs.EventDataTx{tx, nil, ""})
					}
					return nil

				case *txs.UnbondTx:
					// The validator must be active
					_, val := _s.BondedValidators.GetByAddress(tx.Address)
					if val == nil {
						return txs.ErrTxInvalidAddress
					}

					// Verify the signature
					signBytes := acm.SignBytes(_s.ChainID, tx)
					if !val.PubKey.VerifyBytes(signBytes, tx.Signature) {
						return txs.ErrTxInvalidSignature
					}

					// tx.Height must be greater than val.LastCommitHeight
					if tx.Height <= val.LastCommitHeight {
						return errors.New("Invalid unbond height")
					}

					// Good!
					_s.unbondValidator(val)
					if evc != nil {
						evc.FireEvent(txs.EventStringUnbond(), txs.EventDataTx{tx, nil, ""})
					}
					return nil

				case *txs.RebondTx:
					// The validator must be inactive
					_, val := _s.UnbondingValidators.GetByAddress(tx.Address)
					if val == nil {
						return txs.ErrTxInvalidAddress
					}

					// Verify the signature
					signBytes := acm.SignBytes(_s.ChainID, tx)
					if !val.PubKey.VerifyBytes(signBytes, tx.Signature) {
						return txs.ErrTxInvalidSignature
					}

					// tx.Height must be in a suitable range
					minRebondHeight := _s.LastBlockHeight - (validatorTimeoutBlocks / 2)
					maxRebondHeight := _s.LastBlockHeight + 2
					if !((minRebondHeight <= tx.Height) && (tx.Height <= maxRebondHeight)) {
						return errors.New(Fmt("Rebond height not in range.  Expected %v <= %v <= %v",
							minRebondHeight, tx.Height, maxRebondHeight))
					}

					// Good!
					_s.rebondValidator(val)
					if evc != nil {
						evc.FireEvent(txs.EventStringRebond(), txs.EventDataTx{tx, nil, ""})
					}
					return nil

				case *txs.DupeoutTx:
					// Verify the signatures
					_, accused := _s.BondedValidators.GetByAddress(tx.Address)
					if accused == nil {
						_, accused = _s.UnbondingValidators.GetByAddress(tx.Address)
						if accused == nil {
							return txs.ErrTxInvalidAddress
						}
					}
					voteASignBytes := acm.SignBytes(_s.ChainID, &tx.VoteA)
					voteBSignBytes := acm.SignBytes(_s.ChainID, &tx.VoteB)
					if !accused.PubKey.VerifyBytes(voteASignBytes, tx.VoteA.Signature) ||
						!accused.PubKey.VerifyBytes(voteBSignBytes, tx.VoteB.Signature) {
						return txs.ErrTxInvalidSignature
					}

					// Verify equivocation
					// TODO: in the future, just require one vote from a previous height that
					// doesn't exist on this chain.
					if tx.VoteA.Height != tx.VoteB.Height {
						return errors.New("DupeoutTx heights don't match")
					}
					if tx.VoteA.Round != tx.VoteB.Round {
						return errors.New("DupeoutTx rounds don't match")
					}
					if tx.VoteA.Type != tx.VoteB.Type {
						return errors.New("DupeoutTx types don't match")
					}
					if bytes.Equal(tx.VoteA.BlockHash, tx.VoteB.BlockHash) {
						return errors.New("DupeoutTx blockhashes shouldn't match")
					}

					// Good! (Bad validator!)
					_s.destroyValidator(accused)
					if evc != nil {
						evc.FireEvent(txs.EventStringDupeout(), txs.EventDataTx{tx, nil, ""})
					}
					return nil
	*/

	case *txs.PermissionsTx:
		var inAcc *acm.Account

		// Validate input
		inAcc = blockCache.GetAccount(tx.Input.Address)
		if inAcc == nil {
			log.Debug(fmt.Sprintf("Can't find in account %X", tx.Input.Address))
			return txs.ErrTxInvalidAddress
		}

		permFlag := tx.PermArgs.PermFlag()
		// check permission
		if !HasPermission(blockCache, inAcc, permFlag) {
			return fmt.Errorf("Account %X does not have moderator permission %s (%b)", tx.Input.Address, ptypes.PermFlagToString(permFlag), permFlag)
		}

		// pubKey should be present in either "inAcc" or "tx.Input"
		if err := checkInputPubKey(inAcc, tx.Input); err != nil {
			log.Debug(fmt.Sprintf("Can't find pubkey for %X", tx.Input.Address))
			return err
		}
		signBytes := acm.SignBytes(_s.ChainID, tx)
		err := validateInput(inAcc, signBytes, tx.Input)
		if err != nil {
			log.Debug(fmt.Sprintf("validateInput failed on %X: %v", tx.Input.Address, err))
			return err
		}

		value := tx.Input.Amount

		log.Debug("New PermissionsTx", "function", ptypes.PermFlagToString(permFlag), "args", tx.PermArgs)

		var permAcc *acm.Account
		switch args := tx.PermArgs.(type) {
		case *ptypes.HasBaseArgs:
			// this one doesn't make sense from txs
			return fmt.Errorf("HasBase is for contracts, not humans. Just look at the blockchain")
		case *ptypes.SetBaseArgs:
			if permAcc = blockCache.GetAccount(args.Address); permAcc == nil {
				return fmt.Errorf("Trying to update permissions for unknown account %X", args.Address)
			}
			err = permAcc.Permissions.Base.Set(args.Permission, args.Value)
		case *ptypes.UnsetBaseArgs:
			if permAcc = blockCache.GetAccount(args.Address); permAcc == nil {
				return fmt.Errorf("Trying to update permissions for unknown account %X", args.Address)
			}
			err = permAcc.Permissions.Base.Unset(args.Permission)
		case *ptypes.SetGlobalArgs:
			if permAcc = blockCache.GetAccount(ptypes.GlobalPermissionsAddress); permAcc == nil {
				sanity.PanicSanity("can't find global permissions account")
			}
			err = permAcc.Permissions.Base.Set(args.Permission, args.Value)
		case *ptypes.HasRoleArgs:
			return fmt.Errorf("HasRole is for contracts, not humans. Just look at the blockchain")
		case *ptypes.AddRoleArgs:
			if permAcc = blockCache.GetAccount(args.Address); permAcc == nil {
				return fmt.Errorf("Trying to update roles for unknown account %X", args.Address)
			}
			if !permAcc.Permissions.AddRole(args.Role) {
				return fmt.Errorf("Role (%s) already exists for account %X", args.Role, args.Address)
			}
		case *ptypes.RmRoleArgs:
			if permAcc = blockCache.GetAccount(args.Address); permAcc == nil {
				return fmt.Errorf("Trying to update roles for unknown account %X", args.Address)
			}
			if !permAcc.Permissions.RmRole(args.Role) {
				return fmt.Errorf("Role (%s) does not exist for account %X", args.Role, args.Address)
			}
		default:
			sanity.PanicSanity(fmt.Sprintf("invalid permission function: %s", ptypes.PermFlagToString(permFlag)))
		}

		// TODO: maybe we want to take funds on error and allow txs in that don't do anythingi?
		if err != nil {
			return err
		}

		// Good!
		inAcc.Sequence += 1
		inAcc.Balance -= value
		blockCache.UpdateAccount(inAcc)
		if permAcc != nil {
			blockCache.UpdateAccount(permAcc)
		}

		if evc != nil {
			evc.FireEvent(txs.EventStringAccInput(tx.Input.Address), txs.EventDataTx{tx, nil, ""})
			evc.FireEvent(txs.EventStringPermissions(ptypes.PermFlagToString(permFlag)), txs.EventDataTx{tx, nil, ""})
		}

		return nil

	default:
		// binary decoding should not let this happen
		sanity.PanicSanity("Unknown Tx type")
		return nil
	}
}

//---------------------------------------------------------------

// Get permission on an account or fall back to global value
func HasPermission(state AccountGetter, acc *acm.Account, perm ptypes.PermFlag) bool {
	if perm > ptypes.AllPermFlags {
		sanity.PanicSanity("Checking an unknown permission in state should never happen")
	}

	if acc == nil {
		// TODO
		// this needs to fall back to global or do some other specific things
		// eg. a bondAcc may be nil and so can only bond if global bonding is true
	}
	permString := ptypes.PermFlagToString(perm)

	v, err := acc.Permissions.Base.Get(perm)
	if _, ok := err.(ptypes.ErrValueNotSet); ok {
		if state == nil {
			sanity.PanicSanity("All known global permissions should be set!")
		}
		log.Info("Permission for account is not set. Querying GlobalPermissionsAddress", "perm", permString)
		return HasPermission(nil, state.GetAccount(ptypes.GlobalPermissionsAddress), perm)
	} else if v {
		log.Info("Account has permission", "address", fmt.Sprintf("%X", acc.Address), "perm", permString)
	} else {
		log.Info("Account does not have permission", "address", fmt.Sprintf("%X", acc.Address), "perm", permString)
	}
	return v
}

// TODO: for debug log the failed accounts
func hasSendPermission(state AccountGetter, accs map[string]*acm.Account) bool {
	for _, acc := range accs {
		if !HasPermission(state, acc, ptypes.Send) {
			return false
		}
	}
	return true
}

func hasNamePermission(state AccountGetter, acc *acm.Account) bool {
	return HasPermission(state, acc, ptypes.Name)
}

func hasCallPermission(state AccountGetter, acc *acm.Account) bool {
	return HasPermission(state, acc, ptypes.Call)
}

func hasCreateContractPermission(state AccountGetter, acc *acm.Account) bool {
	return HasPermission(state, acc, ptypes.CreateContract)
}

func hasCreateAccountPermission(state AccountGetter, accs map[string]*acm.Account) bool {
	for _, acc := range accs {
		if !HasPermission(state, acc, ptypes.CreateAccount) {
			return false
		}
	}
	return true
}

func hasBondPermission(state AccountGetter, acc *acm.Account) bool {
	return HasPermission(state, acc, ptypes.Bond)
}

func hasBondOrSendPermission(state AccountGetter, accs map[string]*acm.Account) bool {
	for _, acc := range accs {
		if !HasPermission(state, acc, ptypes.Bond) {
			if !HasPermission(state, acc, ptypes.Send) {
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
