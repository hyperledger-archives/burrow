package executors

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs/payload"
)

type SendContext struct {
	Tip         blockchain.TipInfo
	StateWriter state.ReaderWriter
	Logger      *logging.Logger
	tx          *payload.SendTx
}

func (ctx *SendContext) Execute(txe *exec.TxExecution) error {
	var ok bool
	ctx.tx, ok = txe.Envelope.Tx.Payload.(*payload.SendTx)
	if !ok {
		return fmt.Errorf("payload must be NameTx, but is: %v", txe.Envelope.Tx.Payload)
	}
	accounts, err := getInputs(ctx.StateWriter, ctx.tx.Inputs)
	if err != nil {
		return err
	}

	// ensure all inputs have send permissions
	if !hasSendPermission(ctx.StateWriter, accounts, ctx.Logger) {
		return fmt.Errorf("at least one input lacks permission for SendTx")
	}

	// add outputs to accounts map
	// if any outputs don't exist, all inputs must have CreateAccount perm
	accounts, err = getOrMakeOutputs(ctx.StateWriter, accounts, ctx.tx.Outputs, ctx.Logger)
	if err != nil {
		return err
	}

	inTotal, err := validateInputs(accounts, ctx.tx.Inputs)
	if err != nil {
		return err
	}
	outTotal, err := validateOutputs(ctx.tx.Outputs)
	if err != nil {
		return err
	}
	if outTotal > inTotal {
		return errors.ErrorCodeInsufficientFunds
	}
	if outTotal < inTotal {
		return errors.ErrorCodeOverpayment
	}
	if outTotal == 0 {
		return errors.ErrorCodeZeroPayment
	}

	// Good! Adjust accounts
	err = adjustByInputs(accounts, ctx.tx.Inputs, ctx.Logger)
	if err != nil {
		return err
	}

	err = adjustByOutputs(accounts, ctx.tx.Outputs)
	if err != nil {
		return err
	}

	for _, acc := range accounts {
		ctx.StateWriter.UpdateAccount(acc)
	}

	for _, i := range ctx.tx.Inputs {
		txe.Input(i.Address, nil)
	}

	for _, o := range ctx.tx.Outputs {
		txe.Output(o.Address, nil)
	}

	return nil
}
