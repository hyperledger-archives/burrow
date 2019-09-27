package contexts

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/registry"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs/payload"
)

type IdentifyContext struct {
	NodeWriter  registry.ReaderWriter
	StateReader acmstate.Reader
	Logger      *logging.Logger
	tx          *payload.IdentifyTx
}

func (ctx *IdentifyContext) Execute(txe *exec.TxExecution, p payload.Payload) error {
	var ok bool
	ctx.tx, ok = p.(*payload.IdentifyTx)
	if !ok {
		return fmt.Errorf("payload must be IdentifyTx, but is: %v", txe.Envelope.Tx.Payload)
	}

	accounts, _, err := getInputs(ctx.StateReader, ctx.tx.Inputs)
	if err != nil {
		return err
	}

	publicKey := ctx.tx.Node.ValidatorPublicKey
	account, err := ctx.StateReader.GetAccount(publicKey.GetAddress())
	if err != nil {
		return err
	} else if account == nil {
		ctx.Logger.InfoMsg("cannot find account",
			"public_key", publicKey)
		return errors.ErrorCodeInvalidAddress
	}

	if _, ok := accounts[account.GetAddress()]; !ok {
		return fmt.Errorf("target account %s not in tx inputs", account.Address.String())
	}

	// a pre-bonded node must submit on a peers behalf
	err = oneHasPermission(ctx.StateReader, permission.Identify, accounts, ctx.Logger)
	if err != nil {
		return errors.Wrap(err, "at least one input lacks permission for IdentifyTx")
	}

	return ctx.NodeWriter.UpdateNode(account.GetAddress(), ctx.tx.Node)
}
