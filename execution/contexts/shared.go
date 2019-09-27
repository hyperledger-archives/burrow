package contexts

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs/payload"
)

type Context interface {
	Execute(txe *exec.TxExecution, p payload.Payload) error
}

// The accounts from the TxInputs must either already have
// acm.PublicKey().(type) != nil, (it must be known),
// or it must be specified in the TxInput.  If redeclared,
// the TxInput is modified and input.PublicKey() set to nil.
func getInputs(accountGetter acmstate.AccountGetter, ins []*payload.TxInput) (map[crypto.Address]*acm.Account, uint64, error) {
	var total uint64
	accounts := map[crypto.Address]*acm.Account{}
	for _, in := range ins {
		// Account shouldn't be duplicated
		if _, ok := accounts[in.Address]; ok {
			return nil, total, errors.ErrorCodeDuplicateAddress
		}
		acc, err := accountGetter.GetAccount(in.Address)
		if err != nil {
			return nil, total, err
		}
		if acc == nil {
			return nil, total, errors.ErrorCodeInvalidAddress
		}
		accounts[in.Address] = acc
		total += in.Amount
	}
	return accounts, total, nil
}

func getOrMakeOutputs(accountGetter acmstate.AccountGetter, accs map[crypto.Address]*acm.Account,
	outs []*payload.TxOutput, logger *logging.Logger) (map[crypto.Address]*acm.Account, error) {
	if accs == nil {
		accs = make(map[crypto.Address]*acm.Account)
	}
	// we should err if an account is being created but the inputs don't have permission
	var err error
	for _, out := range outs {
		accs[out.Address], err = getOrMakeOutput(accountGetter, accs, out.Address, logger)
		if err != nil {
			return nil, err
		}
	}
	return accs, nil
}

func getOrMakeOutput(accountGetter acmstate.AccountGetter, accs map[crypto.Address]*acm.Account,
	outputAddress crypto.Address, logger *logging.Logger) (*acm.Account, error) {

	// Account shouldn't be duplicated
	if _, ok := accs[outputAddress]; ok {
		return nil, errors.ErrorCodeDuplicateAddress
	}
	acc, err := accountGetter.GetAccount(outputAddress)
	if err != nil {
		return nil, err
	}
	// output account may be nil (new)
	if acc == nil {
		if !hasCreateAccountPermission(accountGetter, accs, logger) {
			return nil, fmt.Errorf("at least one input does not have permission to create accounts")
		}
		logger.InfoMsg("Account not found so attempting to create it", "address", outputAddress)
		acc = &acm.Account{
			Address:     outputAddress,
			Sequence:    0,
			Balance:     0,
			Permissions: permission.ZeroAccountPermissions,
		}
	}

	return acc, nil
}

func validateOutputs(outs []*payload.TxOutput) (uint64, error) {
	total := uint64(0)
	for _, out := range outs {
		// Good. Add amount to total
		total += out.Amount
	}
	return total, nil
}

func adjustByInputs(accs map[crypto.Address]*acm.Account, ins []*payload.TxInput) error {
	for _, in := range ins {
		acc := accs[in.Address]
		if acc == nil {
			return fmt.Errorf("adjustByInputs() expects account in accounts, but account %s not found", in.Address)
		}
		if acc.Balance < in.Amount {
			return fmt.Errorf("adjustByInputs() expects sufficient funds but account %s only has balance %v and "+
				"we are deducting %v", in.Address, acc.Balance, in.Amount)
		}
		err := acc.SubtractFromBalance(in.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func adjustByOutputs(accs map[crypto.Address]*acm.Account, outs []*payload.TxOutput) error {
	for _, out := range outs {
		acc := accs[out.Address]
		if acc == nil {
			return fmt.Errorf("adjustByOutputs() expects account in accounts, but account %s not found",
				out.Address)
		}
		err := acc.AddToBalance(out.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

//---------------------------------------------------------------

// Get permission on an account or fall back to global value
func HasPermission(accountGetter acmstate.AccountGetter, acc *acm.Account, perm permission.PermFlag, logger *logging.Logger) bool {
	if perm > permission.AllPermFlags {
		logger.InfoMsg(
			fmt.Sprintf("HasPermission called on invalid permission 0b%b (invalid) > 0b%b (maximum) ",
				perm, permission.AllPermFlags),
			"invalid_permission", perm,
			"maximum_permission", permission.AllPermFlags)
		return false
	}

	v, err := acc.Permissions.Base.Compose(acmstate.GlobalAccountPermissions(accountGetter).Base).Get(perm)
	if err != nil {
		logger.TraceMsg("Error obtaining permission value (will default to false/deny)",
			"perm_flag", perm.String(),
			structure.ErrorKey, err)
	}

	if v {
		logger.TraceMsg("Account has permission",
			"account_address", acc.GetAddress,
			"perm_flag", perm.String())
	} else {
		logger.TraceMsg("Account does not have permission",
			"account_address", acc.GetAddress,
			"perm_flag", perm.String())
	}
	return v
}

func allHavePermission(accountGetter acmstate.AccountGetter, perm permission.PermFlag,
	accs map[crypto.Address]*acm.Account, logger *logging.Logger) error {
	for _, acc := range accs {
		if !HasPermission(accountGetter, acc, perm, logger) {
			return errors.PermissionDenied{
				Address: acc.Address,
				Perm:    perm,
			}
		}
	}
	return nil
}

func oneHasPermission(accountGetter acmstate.AccountGetter, perm permission.PermFlag,
	accs map[crypto.Address]*acm.Account, logger *logging.Logger) error {
	for _, acc := range accs {
		if HasPermission(accountGetter, acc, perm, logger) {
			return nil
		}
	}
	return errors.PermissionDenied{
		Perm: perm,
	}
}

func hasProposalPermission(accountGetter acmstate.AccountGetter, acc *acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.Proposal, logger)
}

func hasInputPermission(accountGetter acmstate.AccountGetter, acc *acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.Input, logger)
}

func hasBatchPermission(accountGetter acmstate.AccountGetter, acc *acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.Batch, logger)
}

func hasNamePermission(accountGetter acmstate.AccountGetter, acc *acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.Name, logger)
}

func hasCallPermission(accountGetter acmstate.AccountGetter, acc *acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.Call, logger)
}

func hasCreateContractPermission(accountGetter acmstate.AccountGetter, acc *acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.CreateContract, logger)
}

func hasCreateAccountPermission(accountGetter acmstate.AccountGetter, accs map[crypto.Address]*acm.Account,
	logger *logging.Logger) bool {
	for _, acc := range accs {
		if !HasPermission(accountGetter, acc, permission.CreateAccount, logger) {
			return false
		}
	}
	return true
}

func hasBondPermission(accountGetter acmstate.AccountGetter, acc *acm.Account,
	logger *logging.Logger) bool {
	return HasPermission(accountGetter, acc, permission.Bond, logger)
}

func hasBondOrSendPermission(accountGetter acmstate.AccountGetter, accs map[crypto.Address]*acm.Account,
	logger *logging.Logger) bool {
	for _, acc := range accs {
		if !HasPermission(accountGetter, acc, permission.Bond, logger) {
			if !HasPermission(accountGetter, acc, permission.Send, logger) {
				return false
			}
		}
	}
	return true
}
