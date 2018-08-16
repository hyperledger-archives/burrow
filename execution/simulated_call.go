package execution

import (
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/contexts"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
)

// Run a contract's code on an isolated and unpersisted state
// Cannot be used to create new contracts
func CallSim(reader state.Reader, tip bcm.BlockchainInfo, fromAddress, address crypto.Address, data []byte,
	logger *logging.Logger) (*exec.TxExecution, error) {

	cache := state.NewCache(reader)
	exe := contexts.CallContext{
		RunCall:     true,
		StateWriter: cache,
		Tip:         tip,
		Logger:      logger,
	}

	txe := exec.NewTxExecution(txs.Enclose(tip.ChainID(), &payload.CallTx{
		Input: &payload.TxInput{
			Address: fromAddress,
		},
		Address:  &address,
		Data:     data,
		GasLimit: contexts.GasLimit,
	}))
	err := exe.Execute(txe)
	if err != nil {
		return nil, err
	}
	return txe, nil
}

// Run the given code on an isolated and unpersisted state
// Cannot be used to create new contracts.
func CallCodeSim(reader state.Reader, tip bcm.BlockchainInfo, fromAddress, address crypto.Address, code, data []byte,
	logger *logging.Logger) (*exec.TxExecution, error) {

	// Attach code to target account (overwriting target)
	cache := state.NewCache(reader)
	err := cache.UpdateAccount(acm.ConcreteAccount{
		Address: address,
		Code:    code,
	}.MutableAccount())

	if err != nil {
		return nil, err
	}
	return CallSim(cache, tip, fromAddress, address, data, logger)
}
