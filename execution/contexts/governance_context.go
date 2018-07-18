package contexts

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs/payload"
)

type GovernanceContext struct {
	StateWriter  state.ReaderWriter
	ValidatorSet blockchain.ValidatorSet
	Logger       *logging.Logger
	tx           *payload.GovernanceTx
	txe          *exec.TxExecution
}

// GovernanceTx provides a set of TemplateAccounts and GovernanceContext tries to alter the chain state to match the
// specification given
func (ctx *GovernanceContext) Execute(txe *exec.TxExecution) error {
	var ok bool
	ctx.txe = txe
	ctx.tx, ok = txe.Envelope.Tx.Payload.(*payload.GovernanceTx)
	if !ok {
		return fmt.Errorf("payload must be NameTx, but is: %v", txe.Envelope.Tx.Payload)
	}
	accounts, err := getInputs(ctx.StateWriter, ctx.tx.Inputs)
	if err != nil {
		return err
	}

	// ensure all inputs have root permissions
	if !allHavePermission(ctx.StateWriter, permission.Root, accounts, ctx.Logger) {
		return fmt.Errorf("at least one input lacks Root permission needed for GovernanceTx")
	}

	for _, i := range ctx.tx.Inputs {
		txe.Input(i.Address, nil)
	}

	for _, update := range ctx.tx.AccountUpdates {
		if update.Address == nil && update.PublicKey == nil {
			// We do not want to generate a key
			return fmt.Errorf("could not execution GovernanceTx since account template %v contains neither "+
				"address or public key", update)
		}
		if update.PublicKey != nil {
			address := update.PublicKey.Address()
			if update.Address != nil && address != *update.Address {
				return fmt.Errorf("supplied public key %v whose address %v does not match %v provided by"+
					"GovernanceTx", update.PublicKey, address, update.Address)
			}
			update.Address = &address
		}
		if update.PublicKey == nil && update.Power != nil {
			// If we are updating power we will need the key
			return fmt.Errorf("GovernanceTx must be provided with public key when updating validator power")
		}
		account, err := state.GetMutableAccount(ctx.StateWriter, *update.Address)
		if err != nil {
			return err
		}
		if account == nil {
			return fmt.Errorf("account %v not found so cannot update using template %v", update.Address, update)
		}
		governAccountEvent, err := ctx.updateAccount(account, update)
		if err != nil {
			txe.GovernAccount(governAccountEvent, errors.AsException(err))
			return err
		}
		txe.GovernAccount(governAccountEvent, nil)
	}
	return nil
}

func (ctx *GovernanceContext) updateAccount(account *acm.MutableAccount, update *spec.TemplateAccount) (ev *exec.GovernAccountEvent, err error) {
	ev = &exec.GovernAccountEvent{
		AccountUpdate: update,
	}
	if update.Amount != nil {
		err = account.SetBalance(*update.Amount)
		if err != nil {
			return
		}
	}
	if update.NodeAddress != nil {
		// TODO: can we do something useful if provided with a NodeAddress for an account about to become a validator
		// like add it to persistent peers or pre gossip so it gets inbound connections? If so under which circumstances?
	}
	if update.Power != nil {
		if update.PublicKey == nil {
			err = fmt.Errorf("updateAccount should have PublicKey by this point but appears not to for "+
				"template account: %v", update)
			return
		}
		power := new(big.Int).SetUint64(*update.Power)
		if !power.IsInt64() {
			err = fmt.Errorf("power supplied in update to validator power for %v does not fit into int64 and "+
				"so is not supported by Tendermint", update.Address)
		}
		_, err := ctx.ValidatorSet.AlterPower(*update.PublicKey, power)
		if err != nil {
			return ev, err
		}
	}
	perms := account.Permissions()
	if len(update.Permissions) > 0 {
		perms.Base, err = permission.BasePermissionsFromStringList(update.Permissions)
		if err != nil {
			return
		}
	}
	if len(update.Roles) > 0 {
		perms.Roles = update.Roles
	}
	err = account.SetPermissions(perms)
	if err != nil {
		return
	}
	err = ctx.StateWriter.UpdateAccount(account)
	return
}
