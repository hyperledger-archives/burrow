package vmbridge

import (
	"math/big"

	mintcommon "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/events"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/vm"

	common "github.com/ethereum/go-ethereum/common"
	ethvm "github.com/ethereum/go-ethereum/core/vm"
)

func NewVM(appState vm.AppState, params vm.Params, origin mintcommon.Word256, txid []byte, value int64) *VMBridge {
	env := NewEnv(appState, params, origin, value)
	return &VMBridge{
		env:  env,
		txid: txid,
	}
}

type VMBridge struct {
	env  ethvm.Environment
	txid []byte
}

func (vmb *VMBridge) SetFireable(evc events.Fireable) {
	// TODO
}

func (vmb *VMBridge) Call(caller, callee *vm.Account, code, input []byte, value int64, gas *int64) (output []byte, err error) {
	evm := ethvm.NewVm(vmb.env)

	to := vmb.env.Db().GetAccount(common.BytesToAddress(callee.Address.Postfix(20)))
	from := vmb.env.Db().GetAccount(common.BytesToAddress(caller.Address.Postfix(20)))
	gasPrice := int64(10000) //XXX
	contract := ethvm.NewContract(from, to, big.NewInt(value), big.NewInt(*gas), big.NewInt(gasPrice))
	codeAddr := common.BytesToAddress(callee.Address.Postfix(20))
	// NOTE: if it's a create, the codeAddr should be nil!
	contract.SetCallCode(&codeAddr, code)

	defer contract.Finalise()
	output, err = evm.Run(contract, input)
	return
}
