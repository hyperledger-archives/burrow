package executors

import (
	"fmt"

	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
)

type NameContext struct {
	Tip            blockchain.TipInfo
	StateWriter    state.Writer
	EventPublisher event.Publisher
	NameReg        names.NameRegWriter
	Logger         *logging.Logger
	tx             *payload.NameTx
}

func (ctx *NameContext) Execute(txEnv *txs.Envelope) error {
	var ok bool
	ctx.tx, ok = txEnv.Tx.Payload.(*payload.NameTx)
	if !ok {
		return fmt.Errorf("payload must be NameTx, but is: %v", txEnv.Tx.Payload)
	}
	// Validate input
	inAcc, err := state.GetMutableAccount(ctx.StateWriter, ctx.tx.Input.Address)
	if err != nil {
		return err
	}
	if inAcc == nil {
		ctx.Logger.InfoMsg("Cannot find input account",
			"tx_input", ctx.tx.Input)
		return payload.ErrTxInvalidAddress
	}
	// check permission
	if !hasNamePermission(ctx.StateWriter, inAcc, ctx.Logger) {
		return fmt.Errorf("account %s does not have Name permission", ctx.tx.Input.Address)
	}
	err = validateInput(inAcc, ctx.tx.Input)
	if err != nil {
		ctx.Logger.InfoMsg("validateInput failed",
			"tx_input", ctx.tx.Input, structure.ErrorKey, err)
		return err
	}
	if ctx.tx.Input.Amount < ctx.tx.Fee {
		ctx.Logger.InfoMsg("Sender did not send enough to cover the fee",
			"tx_input", ctx.tx.Input)
		return payload.ErrTxInsufficientFunds
	}

	// validate the input strings
	if err := ctx.tx.ValidateStrings(); err != nil {
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
	entry, err := ctx.NameReg.GetNameRegEntry(ctx.tx.Name)
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
			err := ctx.NameReg.RemoveNameRegEntry(entry.Name)
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
			err := ctx.NameReg.UpdateNameRegEntry(entry)
			if err != nil {
				return err
			}
		}
	} else {
		if expiresIn < names.MinNameRegistrationPeriod {
			return fmt.Errorf("Names must be registered for at least %d blocks", names.MinNameRegistrationPeriod)
		}
		// entry does not exist, so create it
		entry = &names.NameRegEntry{
			Name:    ctx.tx.Name,
			Owner:   ctx.tx.Input.Address,
			Data:    ctx.tx.Data,
			Expires: lastBlockHeight + expiresIn,
		}
		ctx.Logger.TraceMsg("Creating NameReg entry",
			"name", entry.Name,
			"expires_in", expiresIn)
		err := ctx.NameReg.UpdateNameRegEntry(entry)
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
	inAcc.IncSequence()
	inAcc, err = inAcc.SubtractFromBalance(value)
	if err != nil {
		return err
	}
	ctx.StateWriter.UpdateAccount(inAcc)

	// TODO: maybe we want to take funds on error and allow txs in that don't do anythingi?

	if ctx.EventPublisher != nil {
		events.PublishAccountInput(ctx.EventPublisher, ctx.tx.Input.Address, txEnv.Tx, nil, nil)
		events.PublishNameReg(ctx.EventPublisher, txEnv.Tx)
	}

	return nil
}
