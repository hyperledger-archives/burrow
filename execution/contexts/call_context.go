package contexts

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/wasm"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs/payload"
)

// TODO: make configurable
const GasLimit = uint64(1000000)

type CallContext struct {
	StateWriter acmstate.ReaderWriter
	RunCall     bool
	Blockchain  Blockchain
	VMOptions   []func(*evm.VM)
	Logger      *logging.Logger
	tx          *payload.CallTx
	txe         *exec.TxExecution
}

func (ctx *CallContext) Execute(txe *exec.TxExecution, p payload.Payload) error {
	var ok bool
	ctx.tx, ok = p.(*payload.CallTx)
	if !ok {
		return fmt.Errorf("payload must be CallTx, but is: %v", p)
	}
	ctx.txe = txe
	inAcc, outAcc, err := ctx.Precheck()
	if err != nil {
		return err
	}
	// That the fee less than the input amount is checked by Precheck to be greater than or equal to fee
	value := ctx.tx.Input.Amount - ctx.tx.Fee

	if ctx.RunCall {
		return ctx.Deliver(inAcc, outAcc, value)
	}
	return ctx.Check(inAcc, value)
}

func (ctx *CallContext) Precheck() (*acm.Account, *acm.Account, error) {
	var outAcc *acm.Account
	// Validate input
	inAcc, err := ctx.StateWriter.GetAccount(ctx.tx.Input.Address)
	if err != nil {
		return nil, nil, err
	}
	if inAcc == nil {
		return nil, nil, errors.ErrorCodef(errors.ErrorCodeInvalidAddress,
			"Cannot find input account: %v", ctx.tx.Input)
	}

	if ctx.tx.Input.Amount < ctx.tx.Fee {
		return nil, nil, errors.ErrorCodef(errors.ErrorCodeInsufficientFunds,
			"Send did not send enough to cover the fee: %v", ctx.tx.Input)
	}

	// Fees are handle by the CallContext, values transfers (i.e. balances) are handled in the VM (or in Check())
	err = inAcc.SubtractFromBalance(ctx.tx.Fee)
	if err != nil {
		return nil, nil, errors.ErrorCodef(errors.ErrorCodeInsufficientFunds,
			"Input account does not have sufficient balance to cover input amount: %v", ctx.tx.Input)
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
		if evm.IsRegisteredNativeContract(*ctx.tx.Address) {
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

func (ctx *CallContext) Check(inAcc *acm.Account, value uint64) error {
	// We do a trial balance subtraction here
	err := inAcc.SubtractFromBalance(value)
	if err != nil {
		return err
	}
	err = ctx.StateWriter.UpdateAccount(inAcc)
	if err != nil {
		return err
	}
	return nil
}

func (ctx *CallContext) Deliver(inAcc, outAcc *acm.Account, value uint64) error {
	createContract := ctx.tx.Address == nil
	// VM call variables
	var (
		gas     uint64         = ctx.tx.GasLimit
		caller  crypto.Address = inAcc.Address
		callee  crypto.Address = crypto.ZeroAddress // initialized below
		code    []byte         = nil
		wcode   []byte         = nil
		ret     []byte         = nil
		txCache                = evm.NewState(ctx.StateWriter, ctx.Blockchain.BlockHash, acmstate.Named("TxCache"))
		params                 = evm.Params{
			BlockHeight: ctx.Blockchain.LastBlockHeight() + 1,
			BlockTime:   ctx.Blockchain.LastBlockTime().Unix(),
			GasLimit:    GasLimit,
		}
	)

	// get or create callee
	if createContract {
		// We already checked for permission
		callee = crypto.NewContractAddress(caller, ctx.txe.TxHash)
		code = ctx.tx.Data
		wcode = ctx.tx.WASM
		txCache.CreateAccount(callee)
		ctx.Logger.TraceMsg("Creating new contract",
			"contract_address", callee,
			"init_code", code)
	} else {
		if outAcc == nil || (len(outAcc.EVMCode) == 0 && len(outAcc.WASMCode) == 0) {
			// if you call an account that doesn't exist
			// or an account with no code then we take fees (sorry pal)
			// NOTE: it's fine to create a contract and call it within one
			// block (sequence number will prevent re-ordering of those txs)
			// but to create with one contract and call with another
			// you have to wait a block to avoid a re-ordering attack
			// that will take your fees
			var exception *errors.Exception
			if outAcc == nil {
				exception = errors.ErrorCodef(errors.ErrorCodeInvalidAddress,
					"CallTx to an address (%v) that does not exist", ctx.tx.Address)
				ctx.Logger.Info.Log(structure.ErrorKey, exception,
					"caller_address", inAcc.GetAddress(),
					"callee_address", ctx.tx.Address)
			} else {
				exception = errors.ErrorCodef(errors.ErrorCodeInvalidAddress,
					"CallTx to an address (%v) that holds no code", ctx.tx.Address)
				ctx.Logger.Info.Log(exception,
					"caller_address", inAcc.GetAddress(),
					"callee_address", ctx.tx.Address)
			}
			ctx.txe.PushError(exception)
			ctx.CallEvents(exception)
			return nil
		}
		callee = outAcc.Address
		code = txCache.GetEVMCode(callee)
		wcode = txCache.GetWASMCode(callee)
		ctx.Logger.TraceMsg("Calling existing contract",
			"contract_address", callee,
			"input", ctx.tx.Data,
			"contract_code", code)
	}
	ctx.Logger.Trace.Log("callee", callee)

	txHash := ctx.txe.Envelope.Tx.Hash()
	logger := ctx.Logger.With(structure.TxHashKey, txHash)
	var exception errors.CodedError
	if wcode != nil {
		if createContract {
			txCache.InitWASMCode(callee, wcode)
		}
		ret, err := wasm.RunWASM(txCache, callee, createContract, wcode, ctx.tx.Data)
		if err != nil {
			ctx.Logger.InfoMsg("Error returned from WASM", "error", err)
			return err
		}
		err = txCache.Sync()
		if err != nil {
			return err
		}
		ctx.txe.Return(ret, ctx.tx.GasLimit-gas)
	} else {
		// EVM
		vmach := evm.NewVM(params, caller, txHash, logger, ctx.VMOptions...)
		ret, exception = vmach.Call(txCache, ctx.txe, caller, callee, code, ctx.tx.Data, value, &gas)
		if exception != nil {
			// Failure. Charge the gas fee. The 'value' was otherwise not transferred.
			ctx.Logger.InfoMsg("Error on execution",
				structure.ErrorKey, exception)

			ctx.txe.PushError(errors.ErrorCodef(exception.ErrorCode(), "call error: %s\nEVM call trace: %s",
				exception.String(), ctx.txe.CallTrace()))
		} else {
			ctx.Logger.TraceMsg("Successful execution")
			if createContract {
				txCache.InitCode(callee, ret)
			}
			err := txCache.Sync()
			if err != nil {
				return err
			}
		}
		ctx.CallEvents(exception)
		ctx.txe.Return(ret, ctx.tx.GasLimit-gas)
	}

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
