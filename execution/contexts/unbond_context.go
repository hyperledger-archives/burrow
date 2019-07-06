package contexts

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs/payload"
)

type UnbondContext struct {
	StateWriter  acmstate.ReaderWriter
	ValidatorSet validator.Alterer
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

	power := new(big.Int).Neg(new(big.Int).SetUint64(ctx.tx.Input.GetAmount()))
	account, err := validateBonding(ctx.StateWriter, ctx.tx.Input, ctx.tx.PublicKey, ctx.Logger)
	if err != nil {
		return err
	}

	err = account.AddToBalance(power.Uint64())
	if err != nil {
		return err
	}

	_, err = ctx.ValidatorSet.AlterPower(account.PublicKey, power)
	if err != nil {
		return err
	}

	return ctx.StateWriter.UpdateAccount(account)
}
