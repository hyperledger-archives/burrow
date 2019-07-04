package contexts

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/acm/validator"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/registry"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs/payload"
)

type IdentifyContext struct {
	NodeWriter   registry.ReaderWriter
	StateReader  acmstate.Reader
	ValidatorSet validator.Reader
	Logger       *logging.Logger
	tx           *payload.IdentifyTx
}

func (ctx *IdentifyContext) Execute(txe *exec.TxExecution, p payload.Payload) error {
	var ok bool
	ctx.tx, ok = p.(*payload.IdentifyTx)
	if !ok {
		return fmt.Errorf("payload must be IdentifyTx, but is: %v", txe.Envelope.Tx.Payload)
	}

	inAcc, err := ctx.StateReader.GetAccount(ctx.tx.Input.Address)
	if err != nil {
		return err
	} else if inAcc == nil {
		ctx.Logger.InfoMsg("Cannot find input account",
			"tx_input", ctx.tx.Input)
		return errors.ErrorCodeInvalidAddress
	}

	// am I a validator?

	// TODO: getIdentity
	if inAcc.GetAddress() != *ctx.tx.Validator.Address {
		return fmt.Errorf("input account and validator address must match")
	}

	power, err := ctx.ValidatorSet.Power(inAcc.GetAddress())
	if err != nil {
		return fmt.Errorf("could not fetch validator ring: %v", err)
	} else if power == nil || power.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("validator does not exist or is not bonded")
	}

	// TODO: check multisig: node & val
	sigs := txe.Envelope.GetSignatories()
	if len(sigs) != 2 {
		return fmt.Errorf("not enough signatures, wanted validator + node")
	}

	return ctx.NodeWriter.RegisterNode(inAcc.GetAddress(), ctx.tx.Node)
}
