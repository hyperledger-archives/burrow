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

type BondContext struct {
	StateWriter  acmstate.ReaderWriter
	ValidatorSet validator.ReaderWriter
	Logger       *logging.Logger
	tx           *payload.BondTx
}

// Execute a BondTx to add or remove a new validator
func (ctx *BondContext) Execute(txe *exec.TxExecution, p payload.Payload) error {
	var ok bool
	ctx.tx, ok = p.(*payload.BondTx)
	if !ok {
		return fmt.Errorf("payload must be BondTx, but is: %v", txe.Envelope.Tx.Payload)
	}

	// the account initiating the bond
	power := new(big.Int).SetUint64(ctx.tx.Input.GetAmount())
	account, err := ctx.StateWriter.GetAccount(ctx.tx.Input.Address)
	if err != nil {
		return err
	}

	// can the account bond?
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

	// we're good to go
	err = account.SubtractFromBalance(amount)
	if err != nil {
		return err
	}

	// assume public key is know as we update account from signatures
	err = validator.AddPower(ctx.ValidatorSet, account.PublicKey, power)
	if err != nil {
		return err
	}

	return ctx.StateWriter.UpdateAccount(account)
}
