package contexts

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs/payload"
)

type GovernanceContext struct {
	StateWriter  state.ReaderWriter
	ValidatorSet validator.Writer
	Logger       *logging.Logger
	tx           *payload.GovTx
	txe          *exec.TxExecution
}

// GovTx provides a set of TemplateAccounts and GovernanceContext tries to alter the chain state to match the
// specification given
func (ctx *GovernanceContext) Execute(txe *exec.TxExecution) error {
	var ok bool
	ctx.txe = txe
	ctx.tx, ok = txe.Envelope.Tx.Payload.(*payload.GovTx)
	if !ok {
		return fmt.Errorf("payload must be NameTx, but is: %v", txe.Envelope.Tx.Payload)
	}
	// Nothing down with any incoming funds at this point
	accounts, _, err := getInputs(ctx.StateWriter, ctx.tx.Inputs)
	if err != nil {
		return err
	}

	// ensure all inputs have root permissions
	err = allHavePermission(ctx.StateWriter, permission.Root, accounts, ctx.Logger)
	if err != nil {
		return errors.Wrap(err, "at least one input lacks permission for GovTx")
	}

	for _, i := range ctx.tx.Inputs {
		txe.Input(i.Address, nil)
	}

	for _, update := range ctx.tx.AccountUpdates {
		if update.Address == nil && update.PublicKey == nil {
			// We do not want to generate a key
			return fmt.Errorf("could not execution GovTx since account template %v contains neither "+
				"address or public key", update)
		}
		if update.PublicKey == nil {
			update.PublicKey, err = ctx.MaybeGetPublicKey(*update.Address)
			if err != nil {
				return err
			}
		}
		// Check address
		if update.PublicKey != nil {
			address := update.PublicKey.Address()
			if update.Address != nil && address != *update.Address {
				return fmt.Errorf("supplied public key %v whose address %v does not match %v provided by"+
					"GovTx", update.PublicKey, address, update.Address)
			}
			update.Address = &address
		} else if update.Balances().HasPower() {
			// If we are updating power we will need the key
			return fmt.Errorf("GovTx must be provided with public key when updating validator power")
		}
		account, err := getOrMakeOutput(ctx.StateWriter, accounts, *update.Address, ctx.Logger)
		if err != nil {
			return err
		}
		governAccountEvent, err := ctx.UpdateAccount(account, update)
		if err != nil {
			txe.GovernAccount(governAccountEvent, errors.AsException(err))
			return err
		}
		txe.GovernAccount(governAccountEvent, nil)
	}
	return nil
}

func (ctx *GovernanceContext) UpdateAccount(account *acm.MutableAccount, update *spec.TemplateAccount) (ev *exec.GovernAccountEvent, err error) {
	ev = &exec.GovernAccountEvent{
		AccountUpdate: update,
	}
	if update.Balances().HasNative() {
		err = account.SetBalance(update.Balances().GetNative(0))
		if err != nil {
			return
		}
	}
	if update.NodeAddress != nil {
		// TODO: can we do something useful if provided with a NodeAddress for an account about to become a validator
		// like add it to persistent peers or pre gossip so it gets inbound connections? If so under which circumstances?
	}
	if update.Balances().HasPower() {
		if update.PublicKey == nil {
			err = fmt.Errorf("updateAccount should have PublicKey by this point but appears not to for "+
				"template account: %v", update)
			return
		}
		power := new(big.Int).SetUint64(update.Balances().GetPower(0))
		if !power.IsInt64() {
			err = fmt.Errorf("power supplied in update to validator power for %v does not fit into int64 and "+
				"so is not supported by Tendermint", update.Address)
		}
		_, err := ctx.ValidatorSet.AlterPower(*update.PublicKey, power)
		if err != nil {
			return ev, err
		}
	}
	if update.Code != nil {
		err = account.SetCode(*update.Code)
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

func (ctx *GovernanceContext) MaybeGetPublicKey(address crypto.Address) (*crypto.PublicKey, error) {
	// First try state in case chain has received input previously
	acc, err := ctx.StateWriter.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if acc != nil && acc.PublicKey().IsSet() {
		publicKey := acc.PublicKey()
		return &publicKey, nil
	}
	return nil, nil
}
