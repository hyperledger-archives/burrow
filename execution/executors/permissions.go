package executors

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
)

type PermissionsContext struct {
	StateWriter    state.Writer
	EventPublisher event.Publisher
	Logger         *logging.Logger
	tx             *payload.PermissionsTx
}

func (ctx *PermissionsContext) Execute(txEnv *txs.Envelope) error {
	var ok bool
	ctx.tx, ok = txEnv.Tx.Payload.(*payload.PermissionsTx)
	if !ok {
		return fmt.Errorf("payload must be PermissionsTx, but is: %v", txEnv.Tx.Payload)
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

	err = ctx.tx.PermArgs.EnsureValid()
	if err != nil {
		return fmt.Errorf("PermissionsTx received containing invalid PermArgs: %v", err)
	}

	permFlag := ctx.tx.PermArgs.PermFlag
	// check permission
	if !HasPermission(ctx.StateWriter, inAcc, permFlag, ctx.Logger) {
		return fmt.Errorf("account %s does not have moderator permission %s (%b)", ctx.tx.Input.Address,
			permFlag.String(), permFlag)
	}

	err = validateInput(inAcc, ctx.tx.Input)
	if err != nil {
		ctx.Logger.InfoMsg("validateInput failed",
			"tx_input", ctx.tx.Input,
			structure.ErrorKey, err)
		return err
	}

	value := ctx.tx.Input.Amount

	ctx.Logger.TraceMsg("New PermissionsTx",
		"perm_args", ctx.tx.PermArgs.String())

	var permAcc acm.Account
	switch ctx.tx.PermArgs.PermFlag {
	case ptypes.HasBase:
		// this one doesn't make sense from txs
		return fmt.Errorf("HasBase is for contracts, not humans. Just look at the blockchain")
	case ptypes.SetBase:
		permAcc, err = mutatePermissions(ctx.StateWriter, *ctx.tx.PermArgs.Address,
			func(perms *ptypes.AccountPermissions) error {
				return perms.Base.Set(*ctx.tx.PermArgs.Permission, *ctx.tx.PermArgs.Value)
			})
	case ptypes.UnsetBase:
		permAcc, err = mutatePermissions(ctx.StateWriter, *ctx.tx.PermArgs.Address,
			func(perms *ptypes.AccountPermissions) error {
				return perms.Base.Unset(*ctx.tx.PermArgs.Permission)
			})
	case ptypes.SetGlobal:
		permAcc, err = mutatePermissions(ctx.StateWriter, acm.GlobalPermissionsAddress,
			func(perms *ptypes.AccountPermissions) error {
				return perms.Base.Set(*ctx.tx.PermArgs.Permission, *ctx.tx.PermArgs.Value)
			})
	case ptypes.HasRole:
		return fmt.Errorf("HasRole is for contracts, not humans. Just look at the blockchain")
	case ptypes.AddRole:
		permAcc, err = mutatePermissions(ctx.StateWriter, *ctx.tx.PermArgs.Address,
			func(perms *ptypes.AccountPermissions) error {
				if !perms.AddRole(*ctx.tx.PermArgs.Role) {
					return fmt.Errorf("role (%s) already exists for account %s",
						*ctx.tx.PermArgs.Role, *ctx.tx.PermArgs.Address)
				}
				return nil
			})
	case ptypes.RemoveRole:
		permAcc, err = mutatePermissions(ctx.StateWriter, *ctx.tx.PermArgs.Address,
			func(perms *ptypes.AccountPermissions) error {
				if !perms.RmRole(*ctx.tx.PermArgs.Role) {
					return fmt.Errorf("role (%s) does not exist for account %s",
						*ctx.tx.PermArgs.Role, *ctx.tx.PermArgs.Address)
				}
				return nil
			})
	default:
		return fmt.Errorf("invalid permission function: %v", permFlag)
	}

	// TODO: maybe we want to take funds on error and allow txs in that don't do anythingi?
	if err != nil {
		return err
	}

	// Good!
	ctx.Logger.TraceMsg("Incrementing sequence number for PermissionsTx",
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
	if permAcc != nil {
		ctx.StateWriter.UpdateAccount(permAcc)
	}

	if ctx.EventPublisher != nil {
		events.PublishAccountInput(ctx.EventPublisher, ctx.tx.Input.Address, txEnv.Tx, nil, nil)
		events.PublishPermissions(ctx.EventPublisher, permFlag, txEnv.Tx)
	}

	return nil
}

func mutatePermissions(stateReader state.Reader, address crypto.Address,
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
