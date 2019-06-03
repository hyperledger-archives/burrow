// Copyright 2019 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package execution

import (
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
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
func CallSim(reader acmstate.Reader, tip bcm.BlockchainInfo, fromAddress, address crypto.Address, data []byte,
	logger *logging.Logger) (*exec.TxExecution, error) {

	cache := acmstate.NewCache(reader)
	exe := contexts.CallContext{
		RunCall:     true,
		StateWriter: cache,
		Blockchain:  tip,
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
	err := exe.Execute(txe, txe.Envelope.Tx.Payload)
	if err != nil {
		return nil, err
	}
	return txe, nil
}

// Run the given code on an isolated and unpersisted state
// Cannot be used to create new contracts.
func CallCodeSim(reader acmstate.Reader, tip bcm.BlockchainInfo, fromAddress, address crypto.Address, code, data []byte,
	logger *logging.Logger) (*exec.TxExecution, error) {

	// Attach code to target account (overwriting target)
	cache := acmstate.NewCache(reader)
	err := cache.UpdateAccount(&acm.Account{
		Address: address,
		Code:    code,
	})

	if err != nil {
		return nil, err
	}
	return CallSim(cache, tip, fromAddress, address, data, logger)
}
