package contexts

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs/payload"
)

type BondContext struct {
	StateWriter  acmstate.ReaderWriter
	ValidatorSet validator.ReaderWriter
	Logger       *logging.Logger
	tx           *payload.BondTx
}

// Execute a BondTx to add a new validator
func (ctx *BondContext) Execute(txe *exec.TxExecution, p payload.Payload) error {
	var ok bool
	ctx.tx, ok = p.(*payload.BondTx)
	if !ok {
		return fmt.Errorf("payload must be BondTx, but is: %v", txe.Envelope.Tx.Payload)
	}

	// the account initiating the bond (may be validator)
	account, err := ctx.StateWriter.GetAccount(ctx.tx.Input.Address)
	if err != nil {
		return err
	}

	// check if account is validator
	power, err := ctx.ValidatorSet.Power(account.Address)
	if err != nil {
		return err
	} else if power != nil && power.Cmp(big.NewInt(0)) == 1 {
		// TODO: something with nodekey
		// ctx.tx.NetAddress
	}

	// account is not validator, can it bond someone?
	if !hasBondPermission(ctx.StateWriter, account, ctx.Logger) {
		return fmt.Errorf("account '%s' lacks bond permission", account.Address)
	}

	// check account has enough to bond
	amount := ctx.tx.Input.GetAmount()
	if amount == 0 {
		return fmt.Errorf("nothing to bond")
	} else if account.Balance < amount {
		return fmt.Errorf("insufficient funds, account %s only has balance %v and "+
			"we are deducting %v", account.Address, account.Balance, amount)
	}

	// ensure pubKey of validator is set
	val := ctx.tx.Validator
	err = GetIdentity(ctx.StateWriter, val)
	if err != nil {
		return fmt.Errorf("BondTx: %v", err)
	}

	// can power be added?
	power = new(big.Int).SetUint64(amount)
	if !power.IsInt64() {
		return fmt.Errorf("power supplied by %v does not fit into int64 and "+
			"so is not supported by Tendermint", account.Address)
	}
	priorPow, err := ctx.ValidatorSet.Power(*val.Address)
	if err != nil {
		return err
	}
	postPow := big.NewInt(0).Add(priorPow, power)
	if !postPow.IsInt64() {
		return fmt.Errorf("power supplied in update to validator power for %v does not fit into int64 and "+
			"so is not supported by Tendermint", *val.Address)
	}

	// create the account if not bonder
	if *val.Address != account.Address {
		valAcc, err := ctx.StateWriter.GetAccount(*val.Address)
		if err != nil {
			return err
		} else if valAcc == nil {
			// validator account doesn't exist
			valAcc = &acm.Account{
				Address:     *val.Address,
				Sequence:    0,
				Balance:     0,
				Permissions: permission.ZeroAccountPermissions,
			}
		}
		// pk must be known later to unbond
		valAcc.PublicKey = *val.PublicKey
		err = ctx.StateWriter.UpdateAccount(valAcc)
		if err != nil {
			return err
		}
	}

	// we're good to go
	err = account.SubtractFromBalance(amount)
	if err != nil {
		return err
	}
	err = validator.AddPower(ctx.ValidatorSet, *val.PublicKey, power)
	if err != nil {
		return err
	}

	return ctx.StateWriter.UpdateAccount(account)
}

type UnbondContext struct {
	StateWriter  acmstate.ReaderWriter
	ValidatorSet validator.ReaderWriter
	Logger       *logging.Logger
	tx           *payload.UnbondTx
}

// Execute an UnbondTx to remove a validator
func (ctx *UnbondContext) Execute(txe *exec.TxExecution, p payload.Payload) error {
	var ok bool
	ctx.tx, ok = p.(*payload.UnbondTx)
	if !ok {
		return fmt.Errorf("payload must be UnbondTx, but is: %v", txe.Envelope.Tx.Payload)
	}

	// the unbonding validator
	sender, err := ctx.StateWriter.GetAccount(ctx.tx.Input.Address)
	if err != nil {
		return err
	}

	var signed bool
	// ensure tx is signed by validator
	for _, sig := range txe.Envelope.GetSignatories() {
		if sender.GetPublicKey().String() == sig.GetPublicKey().String() {
			signed = true
		}
	}
	if !signed {
		return fmt.Errorf("can't unbond, not signed by validator")
	}

	recipient, err := ctx.StateWriter.GetAccount(ctx.tx.Output.Address)
	if err != nil {
		return err
	}

	// make sure that the validator has power to remove
	power, err := ctx.ValidatorSet.Power(sender.Address)
	if err != nil {
		return err
	} else if power == nil || power.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("nothing bonded for validator '%s'", sender.Address)
	}

	publicKey, err := MaybeGetPublicKey(ctx.StateWriter, sender.Address)
	if err != nil {
		return err
	} else if publicKey == nil {
		return fmt.Errorf("need public key to unbond '%s'", sender.Address)
	}

	// remove power and transfer to output
	err = validator.SubtractPower(ctx.ValidatorSet, *publicKey, power)
	if err != nil {
		return err
	}

	err = recipient.AddToBalance(power.Uint64())
	if err != nil {
		return err
	}

	return ctx.StateWriter.UpdateAccount(recipient)
}
