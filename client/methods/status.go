// Copyright 2017 Monax Industries Limited
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

package methods

import (
	"fmt"

	"github.com/hyperledger/burrow/client"
	"github.com/hyperledger/burrow/definitions"
)

func Status(do *definitions.ClientDo) error {
	logger, err := loggerFromClientDo(do, "Status")
	if err != nil {
		return fmt.Errorf("Could not generate logging config from Do: %s", err)
	}
	burrowNodeClient := client.NewBurrowNodeClient(do.NodeAddrFlag, logger)
	genesisHash, validatorPublicKey, latestBlockHash, latestBlockHeight, latestBlockTime, err := burrowNodeClient.Status()
	if err != nil {
		return fmt.Errorf("Error requesting status from chain at (%s): %s", do.NodeAddrFlag, err)
	}

	chainName, chainId, genesisHashfromChainId, err := burrowNodeClient.ChainId()
	if err != nil {
		return fmt.Errorf("Error requesting chainId from chain at (%s): %s", do.NodeAddrFlag, err)
	}

	logger.Info("chain", do.NodeAddrFlag,
		"genesisHash", fmt.Sprintf("%X", genesisHash),
		"chainName", chainName,
		"chainId", chainId,
		"genesisHash from chainId", fmt.Sprintf("%X", genesisHashfromChainId),
		"validator public key", fmt.Sprintf("%X", validatorPublicKey),
		"latest block hash", fmt.Sprintf("%X", latestBlockHash),
		"latest block height", latestBlockHeight,
		"latest block time", latestBlockTime,
	)
	return nil
}
