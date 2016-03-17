package vmbridge

import (
	"math/big"

	mintcommon "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/events"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/vm"

	common "github.com/ethereum/go-ethereum/common"
	ethvm "github.com/ethereum/go-ethereum/core/vm"
)

func NewVM(appState vm.AppState, params vm.Params, origin mintcommon.Word256, txid []byte, value int64) *VMBridge {
	env := NewEnv(appState, params, origin, value)
	return &VMBridge{
		env:    env,
		origin: origin,
		txid:   txid,
	}
}

type VMBridge struct {
	env    ethvm.Environment
	origin mintcommon.Word256
	txid   []byte
	evc    events.Fireable
}

func (vmb *VMBridge) SetFireable(evc events.Fireable) {
	vmb.evc = evc
}

// a bridge call to start the evm. state mechanics have already been dealt with
func (vmb *VMBridge) Call(caller, callee *vm.Account, code, input []byte, value, gas int64) (output []byte, err error) {

	exception := new(string)
	// fire the post call event (including exception if applicable)
	defer vmb.fireCallEvent(exception, &output, caller, callee, input, value, gas)
	// TODO: events for inner calls!

	evm := ethvm.NewVm(vmb.env)

	to := vmb.env.Db().GetAccount(common.BytesToAddress(callee.Address.Postfix(20)))
	from := vmb.env.Db().GetAccount(common.BytesToAddress(caller.Address.Postfix(20)))
	gasPrice := int64(10000) //XXX
	contract := ethvm.NewContract(from, to, big.NewInt(value), big.NewInt(gas), big.NewInt(gasPrice))
	codeAddr := common.BytesToAddress(callee.Address.Postfix(20))

	// proxy for create (tho its possible to have callees with no code, but then we shouldn't care anyways
	if len(callee.Code) == 0 {
		codeAddr = common.Address{}
	}
	contract.SetCallCode(&codeAddr, code) // XXX: why pointer?

	defer contract.Finalise()
	output, err = evm.Run(contract, input)
	return
}

func (vmb *VMBridge) fireCallEvent(exception *string, output *[]byte, caller, callee *vm.Account, input []byte, value int64, gas int64) {
	// fire the post call event (including exception if applicable)
	if vmb.evc != nil {
		vmb.evc.FireEvent(types.EventStringAccCall(callee.Address.Postfix(20)), types.EventDataCall{
			&types.CallData{caller.Address.Postfix(20), callee.Address.Postfix(20), input, value, gas},
			vmb.origin.Postfix(20),
			vmb.txid,
			*output,
			*exception,
		})
	}
}
