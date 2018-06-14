package executors

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
)

// TODO: make configurable
const GasLimit = uint64(1000000)

type CallContext struct {
	Tip            blockchain.TipInfo
	StateWriter    state.Writer
	EventPublisher event.Publisher
	RunCall        bool
	VMOptions      []func(*evm.VM)
	Logger         *logging.Logger
	tx             *payload.CallTx
	txEnv          *txs.Envelope
}

func (ctx *CallContext) Execute(txEnv *txs.Envelope) error {
	var ok bool
	ctx.tx, ok = txEnv.Tx.Payload.(*payload.CallTx)
	if !ok {
		return fmt.Errorf("payload must be CallTx, but is: %v", txEnv.Tx.Payload)
	}
	ctx.txEnv = txEnv
	inAcc, outAcc, err := ctx.Precheck()
	if err != nil {
		return err
	}
	// That the fee less than the input amount is checked by Precheck
	value := ctx.tx.Input.Amount - ctx.tx.Fee

	if ctx.RunCall {
		ctx.Deliver(inAcc, outAcc, value)
	} else {
		ctx.Check(inAcc, value)
	}

	return nil
}

func (ctx *CallContext) Precheck() (*acm.Account, *acm.Account, error) {
	var outAcc *acm.Account
	// Validate input
	inAcc, err := state.GetAccount(ctx.StateWriter, ctx.tx.Input.Address)
	if err != nil {
		return nil, nil, err
	}
	if inAcc == nil {
		ctx.Logger.InfoMsg("Cannot find input account",
			"tx_input", ctx.tx.Input)
		return nil, nil, payload.ErrTxInvalidAddress
	}

	err = validateInput(inAcc, ctx.tx.Input)
	if err != nil {
		ctx.Logger.InfoMsg("validateInput failed",
			"tx_input", ctx.tx.Input, structure.ErrorKey, err)
		return nil, nil, err
	}
	if ctx.tx.Input.Amount < ctx.tx.Fee {
		ctx.Logger.InfoMsg("Sender did not send enough to cover the fee",
			"tx_input", ctx.tx.Input)
		return nil, nil, payload.ErrTxInsufficientFunds
	}

	ctx.Logger.TraceMsg("Incrementing sequence number for CallTx",
		"tag", "sequence",
		"account", inAcc.Address(),
		"old_sequence", inAcc.Sequence(),
		"new_sequence", inAcc.Sequence()+1)

	inAcc.IncSequence()
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
			return nil, nil, fmt.Errorf("attempt to call a native contract at %s, "+
				"but native contracts cannot be called using CallTx. Use a "+
				"contract that calls the native contract or the appropriate tx "+
				"type (eg. PermissionsTx, NameTx)", ctx.tx.Address)
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

func (ctx *CallContext) Check(inAcc *acm.Account, value uint64) error {
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
	return ctx.StateWriter.UpdateAccount(inAcc)
}

func (ctx *CallContext) Deliver(inAcc, outAcc *acm.Account, value uint64) error {
	createContract := ctx.tx.Address == nil
	// VM call variables
	var (
		gas     uint64       = ctx.tx.GasLimit
		caller  *acm.Account = inAcc
		callee  *acm.Account = nil // initialized below
		code    []byte       = nil
		ret     []byte       = nil
		txCache              = state.NewCache(ctx.StateWriter, state.Name("TxCache"))
		params               = evm.Params{
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
			ctx.FireCallEvents(nil, payload.ErrTxInvalidAddress)
			return nil
		}
		callee = outAcc
		code = callee.Code()
		ctx.Logger.TraceMsg("Calling existing contract",
			"contract_address", callee.Address(),
			"input", ctx.tx.Data,
			"contract_code", code)
	}
	ctx.Logger.Trace.Log("callee", callee.Address().String())

	txCache.UpdateAccount(caller)
	txCache.UpdateAccount(callee)
	vmach := evm.NewVM(params, caller.Address(), ctx.txEnv.Tx.Hash(), ctx.Logger, ctx.VMOptions...)
	vmach.SetPublisher(ctx.EventPublisher)
	// NOTE: Call() transfers the value from caller to callee iff call succeeds.
	ret, exception := vmach.Call(txCache, caller, callee, code, ctx.tx.Data, value, &gas)
	if exception != nil {
		// Failure. Charge the gas fee. The 'value' was otherwise not transferred.
		ctx.Logger.InfoMsg("Error on execution",
			structure.ErrorKey, exception)
	} else {
		ctx.Logger.TraceMsg("Successful execution")
		if createContract {
			callee.SetCode(ret)
		}
		// Update caller/callee to txCache.
		txCache.UpdateAccount(caller)
		txCache.UpdateAccount(callee)

		err := txCache.Sync(ctx.StateWriter)
		if err != nil {
			return err
		}
	}
	// Create a receipt from the ret and whether it erred.
	ctx.Logger.TraceMsg("VM call complete",
		"caller", caller,
		"callee", callee,
		"return", ret,
		structure.ErrorKey, exception)
	ctx.FireCallEvents(ret, exception)
	return nil
}

func (ctx *CallContext) FireCallEvents(ret []byte, err error) {
	// Fire Events for sender and receiver
	// a separate event will be fired from vm for each additional call
	if ctx.EventPublisher != nil {
		events.PublishAccountInput(ctx.EventPublisher, ctx.tx.Input.Address, ctx.txEnv.Tx, ret, errors.AsCodedError(err))
		if ctx.tx.Address != nil {
			events.PublishAccountOutput(ctx.EventPublisher, *ctx.tx.Address, ctx.txEnv.Tx, ret, errors.AsCodedError(err))
		}
	}
}
