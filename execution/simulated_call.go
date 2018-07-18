package execution

import (
	"fmt"
	"runtime/debug"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/contexts"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
)

// Run a contract's code on an isolated and unpersisted state
// Cannot be used to create new contracts
func CallSim(reader state.Reader, tip blockchain.TipInfo, fromAddress, address crypto.Address, data []byte,
	logger *logging.Logger) (*exec.TxExecution, error) {

	if evm.IsRegisteredNativeContract(address.Word256()) {
		return nil, fmt.Errorf("attempt to call native contract at address "+
			"%X, but native contracts can not be called directly. Use a deployed "+
			"contract that calls the native function instead", address)
	}
	// This was being run against CheckTx cache, need to understand the reasoning
	callee, err := state.GetMutableAccount(reader, address)
	if err != nil {
		return nil, err
	}
	if callee == nil {
		return nil, fmt.Errorf("account %s does not exist", address)
	}
	return CallCodeSim(reader, tip, fromAddress, address, callee.Code(), data, logger)
}

// Run the given code on an isolated and unpersisted state
// Cannot be used to create new contracts.
func CallCodeSim(reader state.Reader, tip blockchain.TipInfo, fromAddress, address crypto.Address, code, data []byte,
	logger *logging.Logger) (_ *exec.TxExecution, err error) {
	// This was being run against CheckTx cache, need to understand the reasoning
	caller := acm.ConcreteAccount{Address: fromAddress}.MutableAccount()
	callee := acm.ConcreteAccount{Address: address}.MutableAccount()

	txCache := state.NewCache(reader)

	params := vmParams(tip)
	vmach := evm.NewVM(params, caller.Address(), nil, logger.WithScope("CallCode"))

	txe := &exec.TxExecution{}
	vmach.SetEventSink(txe)
	gas := params.GasLimit
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic from VM in simulated call: %v\n%s", r, debug.Stack())
		}
	}()
	ret, err := vmach.Call(txCache, caller, callee, code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	txe.Return(ret, params.GasLimit-gas)
	return txe, nil
}

func vmParams(tip blockchain.TipInfo) evm.Params {
	return evm.Params{
		BlockHeight: tip.LastBlockHeight(),
		BlockHash:   binary.LeftPadWord256(tip.LastBlockHash()),
		BlockTime:   tip.LastBlockTime().Unix(),
		GasLimit:    contexts.GasLimit,
	}
}
