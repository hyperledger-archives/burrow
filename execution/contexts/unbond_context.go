package contexts

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/hyperledger/burrow/util"
)

type UnbondContext struct {
	State        acmstate.ReaderWriter
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

	if ctx.tx.Input.Address != ctx.tx.Output.Address {
		return fmt.Errorf("input and output address must match")
	}

	power := new(big.Int).SetUint64(ctx.tx.Output.GetAmount())
	account, err := ctx.State.GetAccount(ctx.tx.Input.Address)
	if err != nil {
		return err
	}

	err = account.AddToBalance(power.Uint64())
	if err != nil {
		return err
	}

	util.Debugf("unbonding %v", power)
	cache := ctx.ValidatorSet.(*validator.Cache)
	util.Debugf("%v", cache.Bucket)
	err = validator.SubtractPower(ctx.ValidatorSet, account.PublicKey, power)
	if err != nil {
		return err
	}

	return ctx.State.UpdateAccount(account)
}
