package contexts

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/native"
	"github.com/hyperledger/burrow/execution/wasm"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs/payload"
)

// TODO: make configurable
const GasLimit = uint64(1000000)

type CallContext struct {
	EVM           *evm.EVM
	State         acmstate.ReaderWriter
	MetadataState acmstate.MetadataReaderWriter
	Blockchain    engine.Blockchain
	RunCall       bool
	Logger        *logging.Logger
	tx            *payload.CallTx
	txe           *exec.TxExecution
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
	inAcc, err := ctx.State.GetAccount(ctx.tx.Input.Address)
	if err != nil {
		return nil, nil, err
	}
	if inAcc == nil {
		return nil, nil, errors.Errorf(errors.Codes.InvalidAddress,
			"Cannot find input account: %v", ctx.tx.Input)
	}

	if ctx.tx.Input.Amount < ctx.tx.Fee {
		return nil, nil, errors.Errorf(errors.Codes.InsufficientFunds,
			"Send did not send enough to cover the fee: %v", ctx.tx.Input)
	}

	// Fees are handle by the CallContext, values transfers (i.e. balances) are handled in the VM (or in Check())
	err = inAcc.SubtractFromBalance(ctx.tx.Fee)
	if err != nil {
		return nil, nil, errors.Errorf(errors.Codes.InsufficientFunds,
			"Input account does not have sufficient balance to cover input amount: %v", ctx.tx.Input)
	}

	// Calling a nil destination is defined as requesting contract creation
	createContract := ctx.tx.Address == nil

	if createContract {
		if !hasCreateContractPermission(ctx.State, inAcc, ctx.Logger) {
			return nil, nil, fmt.Errorf("account %s does not have CreateContract permission", ctx.tx.Input.Address)
		}
	} else {
		if !hasCallPermission(ctx.State, inAcc, ctx.Logger) {
			return nil, nil, fmt.Errorf("account %s does not have Call permission", ctx.tx.Input.Address)
		}

		// Output account may be nil if we are still in mempool and contract was created in same block as this tx
		// but that's fine, because the account will be created properly when the create tx runs in the block
		// and then this won't return nil. otherwise, we take their fee
		// Note: ctx.tx.Address == nil iff createContract so dereference is okay
		outAcc, err = ctx.State.GetAccount(*ctx.tx.Address)
		if err != nil {
			return nil, nil, err
		}
	}

	err = ctx.State.UpdateAccount(inAcc)
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
	err = ctx.State.UpdateAccount(inAcc)
	if err != nil {
		return err
	}
	return nil
}

func (ctx *CallContext) Deliver(inAcc, outAcc *acm.Account, value uint64) error {
	// VM call variables
	createContract := ctx.tx.Address == nil
	caller := inAcc.Address
	txCache := acmstate.NewCache(ctx.State, acmstate.Named("TxCache"))
	metaCache := acmstate.NewMetadataCache(ctx.MetadataState)

	var callee crypto.Address
	var code []byte
	var wcode []byte

	// get or create callee
	if createContract {
		// We already checked for permission
		callee = crypto.NewContractAddress(caller, ctx.txe.TxHash)
		code = ctx.tx.Data
		wcode = ctx.tx.WASM
		err := native.CreateAccount(txCache, callee)
		if err != nil {
			return err
		}
		ctx.Logger.TraceMsg("Creating new contract",
			"contract_address", callee,
			"init_code", code)

		// store abis
		err = native.UpdateContractMeta(txCache, metaCache, callee, ctx.tx.ContractMeta)
		if err != nil {
			return err
		}
	} else {
		if outAcc == nil {
			// if you call an account that doesn't exist
			// or an account with no code then we take fees (sorry pal)
			// NOTE: it's fine to create a contract and call it within one
			// block (sequence number will prevent re-ordering of those txs)
			// but to create with one contract and call with another
			// you have to wait a block to avoid a re-ordering attack
			// that will take your fees
			exception := errors.Errorf(errors.Codes.InvalidAddress,
				"CallTx to an address (%v) that does not exist", ctx.tx.Address)
			ctx.Logger.Info.Log(structure.ErrorKey, exception,
				"caller_address", inAcc.GetAddress(),
				"callee_address", ctx.tx.Address)
			ctx.txe.PushError(exception)
			ctx.CallEvents(exception)
			return nil
		}
		callee = outAcc.Address
		acc, err := txCache.GetAccount(callee)
		if err != nil {
			return err
		}
		code = acc.EVMCode
		wcode = acc.WASMCode
		ctx.Logger.TraceMsg("Calling existing contract",
			"contract_address", callee,
			"input", ctx.tx.Data,
			"evm_code", code,
			"wasm_code", wcode)
	}
	ctx.Logger.Trace.Log("callee", callee)

	var ret []byte
	var err error
	txHash := ctx.txe.Envelope.Tx.Hash()
	gas := ctx.tx.GasLimit
	if len(wcode) != 0 {
		ret, err = wasm.RunWASM(txCache, callee, createContract, wcode, ctx.tx.Data)
		if err != nil {
			// Failure. Charge the gas fee. The 'value' was otherwise not transferred.
			ctx.Logger.InfoMsg("Error on WASM execution",
				structure.ErrorKey, err)

			ctx.txe.PushError(errors.Wrap(err, "call error"))
		} else {
			ctx.Logger.TraceMsg("Successful execution")
			if createContract {
				err := native.InitWASMCode(txCache, callee, ret)
				if err != nil {
					return err
				}
			}
			err = ctx.Sync(txCache, metaCache)
			if err != nil {
				return err
			}
		}
	} else {
		// EVM
		ctx.EVM.SetNonce(txHash)
		ctx.EVM.SetLogger(ctx.Logger.With(structure.TxHashKey, txHash))

		params := engine.CallParams{
			Origin: caller,
			Caller: caller,
			Callee: callee,
			Input:  ctx.tx.Data,
			Value:  value,
			Gas:    &gas,
		}

		ret, err = ctx.EVM.Execute(txCache, ctx.Blockchain, ctx.txe, params, code)

		if err != nil {
			// Failure. Charge the gas fee. The 'value' was otherwise not transferred.
			ctx.Logger.InfoMsg("Error on EVM execution",
				structure.ErrorKey, err)

			ctx.txe.PushError(errors.Wrapf(err, "call error: %v\nEVM call trace: %s",
				err, ctx.txe.CallTrace()))
		} else {
			ctx.Logger.TraceMsg("Successful execution")
			if createContract {
				err := native.InitEVMCode(txCache, callee, ret)
				if err != nil {
					return err
				}
			}
			err = ctx.Sync(txCache, metaCache)
			if err != nil {
				return err
			}
		}
		ctx.CallEvents(err)
	}
	ctx.txe.Return(ret, ctx.tx.GasLimit-gas)
	// Create a receipt from the ret and whether it erred.
	ctx.Logger.TraceMsg("VM Call complete",
		"caller", caller,
		"callee", callee,
		"return", ret,
		structure.ErrorKey, err)

	return nil
}

func (ctx *CallContext) CallEvents(err error) {
	// Fire Events for sender and receiver a separate event will be fired from vm for each additional call
	ctx.txe.Input(ctx.tx.Input.Address, errors.AsException(err))
	if ctx.tx.Address != nil {
		ctx.txe.Input(*ctx.tx.Address, errors.AsException(err))
	}
}

func (ctx *CallContext) Sync(cache *acmstate.Cache, metaCache *acmstate.MetadataCache) error {
	err := cache.Sync(ctx.State)
	if err != nil {
		return err
	}
	return metaCache.Sync(ctx.MetadataState)
}
