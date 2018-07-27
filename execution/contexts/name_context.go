package contexts

import (
	"fmt"

	"regexp"

	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs/payload"
)

// Name should be file system like
// Data should be anything permitted in JSON
var regexpAlphaNum = regexp.MustCompile("^[a-zA-Z0-9._/-@]*$")
var regexpJSON = regexp.MustCompile(`^[a-zA-Z0-9_/ \-+"':,\n\t.{}()\[\]]*$`)

type NameContext struct {
	Tip         bcm.BlockchainInfo
	StateWriter state.ReaderWriter
	NameReg     names.ReaderWriter
	Logger      *logging.Logger
	tx          *payload.NameTx
}

func (ctx *NameContext) Execute(txe *exec.TxExecution) error {
	var ok bool
	ctx.tx, ok = txe.Envelope.Tx.Payload.(*payload.NameTx)
	if !ok {
		return fmt.Errorf("payload must be NameTx, but is: %v", txe.Envelope.Tx.Payload)
	}
	// Validate input
	inAcc, err := state.GetMutableAccount(ctx.StateWriter, ctx.tx.Input.Address)
	if err != nil {
		return err
	}
	if inAcc == nil {
		ctx.Logger.InfoMsg("Cannot find input account",
			"tx_input", ctx.tx.Input)
		return errors.ErrorCodeInvalidAddress
	}
	// check permission
	if !hasNamePermission(ctx.StateWriter, inAcc, ctx.Logger) {
		return fmt.Errorf("account %s does not have Name permission", ctx.tx.Input.Address)
	}
	if ctx.tx.Input.Amount < ctx.tx.Fee {
		ctx.Logger.InfoMsg("Sender did not send enough to cover the fee",
			"tx_input", ctx.tx.Input)
		return errors.ErrorCodeInsufficientFunds
	}

	// validate the input strings
	if err := validateStrings(ctx.tx); err != nil {
		return err
	}

	value := ctx.tx.Input.Amount - ctx.tx.Fee

	// let's say cost of a name for one block is len(data) + 32
	costPerBlock := names.NameCostPerBlock(names.NameBaseCost(ctx.tx.Name, ctx.tx.Data))
	expiresIn := value / uint64(costPerBlock)
	lastBlockHeight := ctx.Tip.LastBlockHeight()

	ctx.Logger.TraceMsg("New NameTx",
		"value", value,
		"cost_per_block", costPerBlock,
		"expires_in", expiresIn,
		"last_block_height", lastBlockHeight)

	// check if the name exists
	entry, err := ctx.NameReg.GetName(ctx.tx.Name)
	if err != nil {
		return err
	}

	if entry != nil {
		var expired bool

		// if the entry already exists, and hasn't expired, we must be owner
		if entry.Expires > lastBlockHeight {
			// ensure we are owner
			if entry.Owner != ctx.tx.Input.Address {
				return fmt.Errorf("permission denied: sender %s is trying to update a name (%s) for "+
					"which they are not an owner", ctx.tx.Input.Address, ctx.tx.Name)
			}
		} else {
			expired = true
		}

		// no value and empty data means delete the entry
		if value == 0 && len(ctx.tx.Data) == 0 {
			// maybe we reward you for telling us we can delete this crap
			// (owners if not expired, anyone if expired)
			ctx.Logger.TraceMsg("Removing NameReg entry (no value and empty data in tx requests this)",
				"name", entry.Name)
			err := ctx.NameReg.RemoveName(entry.Name)
			if err != nil {
				return err
			}
		} else {
			// update the entry by bumping the expiry
			// and changing the data
			if expired {
				if expiresIn < names.MinNameRegistrationPeriod {
					return fmt.Errorf("Names must be registered for at least %d blocks", names.MinNameRegistrationPeriod)
				}
				entry.Expires = lastBlockHeight + expiresIn
				entry.Owner = ctx.tx.Input.Address
				ctx.Logger.TraceMsg("An old NameReg entry has expired and been reclaimed",
					"name", entry.Name,
					"expires_in", expiresIn,
					"owner", entry.Owner)
			} else {
				// since the size of the data may have changed
				// we use the total amount of "credit"
				oldCredit := (entry.Expires - lastBlockHeight) * names.NameBaseCost(entry.Name, entry.Data)
				credit := oldCredit + value
				expiresIn = uint64(credit / costPerBlock)
				if expiresIn < names.MinNameRegistrationPeriod {
					return fmt.Errorf("names must be registered for at least %d blocks", names.MinNameRegistrationPeriod)
				}
				entry.Expires = lastBlockHeight + expiresIn
				ctx.Logger.TraceMsg("Updated NameReg entry",
					"name", entry.Name,
					"expires_in", expiresIn,
					"old_credit", oldCredit,
					"value", value,
					"credit", credit)
			}
			entry.Data = ctx.tx.Data
			err := ctx.NameReg.UpdateName(entry)
			if err != nil {
				return err
			}
		}
	} else {
		if expiresIn < names.MinNameRegistrationPeriod {
			return fmt.Errorf("Names must be registered for at least %d blocks", names.MinNameRegistrationPeriod)
		}
		// entry does not exist, so create it
		entry = &names.Entry{
			Name:    ctx.tx.Name,
			Owner:   ctx.tx.Input.Address,
			Data:    ctx.tx.Data,
			Expires: lastBlockHeight + expiresIn,
		}
		ctx.Logger.TraceMsg("Creating NameReg entry",
			"name", entry.Name,
			"expires_in", expiresIn)
		err := ctx.NameReg.UpdateName(entry)
		if err != nil {
			return err
		}
	}

	// TODO: something with the value sent?

	// Good!
	ctx.Logger.TraceMsg("Incrementing sequence number for NameTx",
		"tag", "sequence",
		"account", inAcc.Address(),
		"old_sequence", inAcc.Sequence(),
		"new_sequence", inAcc.Sequence()+1)
	err = inAcc.SubtractFromBalance(value)
	if err != nil {
		return err
	}
	err = ctx.StateWriter.UpdateAccount(inAcc)
	if err != nil {
		return err
	}

	// TODO: maybe we want to take funds on error and allow txs in that don't do anythingi?

	txe.Input(ctx.tx.Input.Address, nil)
	txe.Name(entry)
	return nil
}

func validateStrings(tx *payload.NameTx) error {
	if len(tx.Name) == 0 {
		return errors.ErrorCodef(errors.ErrorCodeInvalidString, "name must not be empty")
	}
	if len(tx.Name) > names.MaxNameLength {
		return errors.ErrorCodef(errors.ErrorCodeInvalidString, "Name is too long. Max %d bytes", names.MaxNameLength)
	}
	if len(tx.Data) > names.MaxDataLength {
		return errors.ErrorCodef(errors.ErrorCodeInvalidString, "Data is too long. Max %d bytes", names.MaxDataLength)
	}

	if !validateNameRegEntryName(tx.Name) {
		return errors.ErrorCodef(errors.ErrorCodeInvalidString,
			"Invalid characters found in NameTx.Name (%s). Only alphanumeric, underscores, dashes, forward slashes, and @ are allowed", tx.Name)
	}

	if !validateNameRegEntryData(tx.Data) {
		return errors.ErrorCodef(errors.ErrorCodeInvalidString,
			"Invalid characters found in NameTx.Data (%s). Only the kind of things found in a JSON file are allowed", tx.Data)
	}

	return nil
}

// filter strings
func validateNameRegEntryName(name string) bool {
	return regexpAlphaNum.Match([]byte(name))
}

func validateNameRegEntryData(data string) bool {
	return regexpJSON.Match([]byte(data))
}
