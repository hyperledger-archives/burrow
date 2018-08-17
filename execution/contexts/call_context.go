package contexts

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs/payload"
)

// TODO: make configurable
const GasLimit = uint64(1000000)

type CallContext struct {
	Tip         bcm.BlockchainInfo
	StateWriter state.ReaderWriter
	RunCall     bool
	VMOptions   []func(*evm.VM)
	Logger      *logging.Logger
	tx          *payload.CallTx
	txe         *exec.TxExecution
}

func (ctx *CallContext) Execute(txe *exec.TxExecution) error {
	var ok bool
	ctx.tx, ok = txe.Envelope.Tx.Payload.(*payload.CallTx)
	if !ok {
		return fmt.Errorf("payload must be CallTx, but is: %v", txe.Envelope.Tx.Payload)
	}
	ctx.txe = txe
	inAcc, outAcc, err := ctx.Precheck()
	if err != nil {
		return err
	}
	// That the fee less than the input amount is checked by Precheck
	value := ctx.tx.Input.Amount - ctx.tx.Fee

	if ctx.RunCall {
		return ctx.Deliver(inAcc, outAcc, value)
	}
	return ctx.Check(inAcc, value)
}

func (ctx *CallContext) Precheck() (*acm.MutableAccount, acm.Account, error) {
	var outAcc acm.Account
	// Validate input
	inAcc, err := state.GetMutableAccount(ctx.StateWriter, ctx.tx.Input.Address)
	if err != nil {
		return nil, nil, err
	}
	if inAcc == nil {
		ctx.Logger.InfoMsg("Cannot find input account",
			"tx_input", ctx.tx.Input)
		return nil, nil, errors.ErrorCodeInvalidAddress
	}

	if ctx.tx.Input.Amount < ctx.tx.Fee {
		ctx.Logger.InfoMsg("Sender did not send enough to cover the fee",
			"tx_input", ctx.tx.Input)
		return nil, nil, errors.ErrorCodeInsufficientFunds
	}

	err = inAcc.SubtractFromBalance(ctx.tx.Fee)
	if err != nil {
		return nil, nil, err
	}

	// Calling a nil destination is defined as requesting contract creation
	createContract := ctx.tx.Address == nil

	if createContract {
		if !hasCreateContractPermission(ctx.StateWriter, inAcc, ctx.Logger) {
			return nil, nil, fmt.Errorf("account %s does not have CreateContract permission", ctx.tx.Input.Address)
		}
	} else {
		if !hasCallPermission(ctx.StateWriter, inAcc, ctx.Logger) {
			return nil, nil, fmt.Errorf("account %s does not have Call permission", ctx.tx.Input.Address)
		}
		// check if its a native contract
		if evm.IsRegisteredNativeContract(ctx.tx.Address.Word256()) {
			return nil, nil, errors.ErrorCodef(errors.ErrorCodeReservedAddress,
				"attempt to call a native contract at %s, "+
					"but native contracts cannot be called using CallTx. Use a "+
					"contract that calls the native contract or the appropriate tx "+
					"type (eg. PermsTx, NameTx)", ctx.tx.Address)
		}

		// Output account may be nil if we are still in mempool and contract was created in same block as this tx
		// but that's fine, because the account will be created properly when the create tx runs in the block
		// and then this won't return nil. otherwise, we take their fee
		// Note: ctx.tx.Address == nil iff createContract so dereference is okay
		outAcc, err = ctx.StateWriter.GetAccount(*ctx.tx.Address)
		if err != nil {
			return nil, nil, err
		}
	}

	err = ctx.StateWriter.UpdateAccount(inAcc)
	if err != nil {
		return nil, nil, err
	}
	return inAcc, outAcc, nil
}

func (ctx *CallContext) Check(inAcc *acm.MutableAccount, value uint64) error {
	createContract := ctx.tx.Address == nil
	// The mempool does not call txs until
	// the proposer determines the order of txs.
	// So mempool will skip the actual .Call(),
	// and only deduct from the caller's balance.
	err := inAcc.SubtractFromBalance(value)
	if err != nil {
		return err
	}
	if createContract {
		// This is done by DeriveNewAccount when runCall == true
		ctx.Logger.TraceMsg("Incrementing sequence number since creates contract",
			"tag", "sequence",
			"account", inAcc.Address(),
			"old_sequence", inAcc.Sequence(),
			"new_sequence", inAcc.Sequence()+1)
		inAcc.IncSequence()
	}
	err = ctx.StateWriter.UpdateAccount(inAcc)
	if err != nil {
		return err
	}
	return nil
}

func (ctx *CallContext) Deliver(inAcc, outAcc acm.Account, value uint64) error {
	createContract := ctx.tx.Address == nil
	// VM call variables
	var (
		gas     uint64              = ctx.tx.GasLimit
		caller  *acm.MutableAccount = acm.AsMutableAccount(inAcc)
		callee  *acm.MutableAccount = nil // initialized below
		code    []byte              = nil
		ret     []byte              = nil
		txCache                     = state.NewCache(ctx.StateWriter, state.Name("TxCache"))
		params                      = evm.Params{
			BlockHeight: ctx.Tip.LastBlockHeight(),
			BlockHash:   binary.LeftPadWord256(ctx.Tip.LastBlockHash()),
			BlockTime:   ctx.Tip.LastBlockTime().Unix(),
			GasLimit:    GasLimit,
		}
	)

	// get or create callee
	if createContract {
		// We already checked for permission
		callee = evm.DeriveNewAccount(caller, state.GlobalAccountPermissions(ctx.StateWriter), ctx.Logger)
		code = ctx.tx.Data
		ctx.Logger.TraceMsg("Creating new contract",
			"contract_address", callee.Address(),
			"init_code", code)
	} else {
		if outAcc == nil || len(outAcc.Code()) == 0 {
			// if you call an account that doesn't exist
			// or an account with no code then we take fees (sorry pal)
			// NOTE: it's fine to create a contract and call it within one
			// block (sequence number will prevent re-ordering of those txs)
			// but to create with one contract and call with another
			// you have to wait a block to avoid a re-ordering attack
			// that will take your fees
			if outAcc == nil {
				ctx.Logger.InfoMsg("Call to address that does not exist",
					"caller_address", inAcc.Address(),
					"callee_address", ctx.tx.Address)
			} else {
				ctx.Logger.InfoMsg("Call to address that holds no code",
					"caller_address", inAcc.Address(),
					"callee_address", ctx.tx.Address)
			}
			ctx.CallEvents(errors.ErrorCodeInvalidAddress)
			return nil
		}
		callee = acm.AsMutableAccount(outAcc)
		code = callee.Code()
		ctx.Logger.TraceMsg("Calling existing contract",
			"contract_address", callee.Address(),
			"input", ctx.tx.Data,
			"contract_code", code)
	}
	ctx.Logger.Trace.Log("callee", callee.Address().String())

	txCache.UpdateAccount(caller)
	txCache.UpdateAccount(callee)
	vmach := evm.NewVM(params, caller.Address(), ctx.txe.Envelope.Tx, ctx.Logger, ctx.VMOptions...)
	vmach.SetEventSink(ctx.txe)
	// NOTE: Call() transfers the value from caller to callee iff call succeeds.
	ret, exception := vmach.Call(txCache, caller, callee, code, ctx.tx.Data, value, &gas)
	if exception != nil {
		// Failure. Charge the gas fee. The 'value' was otherwise not transferred.
		ctx.Logger.InfoMsg("Error on execution",
			structure.ErrorKey, exception)
		ctx.txe.SetException(exception)
	} else {
		ctx.Logger.TraceMsg("Successful execution")
		if createContract {
			callee.SetCode(ret)
		}
		err := txCache.Sync(ctx.StateWriter)
		if err != nil {
			return err
		}
	}
	ctx.CallEvents(exception)
	ctx.txe.Return(ret, ctx.tx.GasLimit-gas)
	// Create a receipt from the ret and whether it erred.
	ctx.Logger.TraceMsg("VM call complete",
		"caller", caller,
		"callee", callee,
		"return", ret,
		structure.ErrorKey, exception)
	return nil
}

func (ctx *CallContext) CallEvents(err error) {
	// Fire Events for sender and receiver a separate event will be fired from vm for each additional call
	ctx.txe.Input(ctx.tx.Input.Address, errors.AsException(err))
	if ctx.tx.Address != nil {
		ctx.txe.Input(*ctx.tx.Address, errors.AsException(err))
	}
}
