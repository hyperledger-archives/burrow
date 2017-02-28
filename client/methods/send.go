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

	"github.com/eris-ltd/eris-db/client"
	"github.com/eris-ltd/eris-db/client/rpc"
	"github.com/eris-ltd/eris-db/definitions"
	"github.com/eris-ltd/eris-db/keys"
)

func Send(do *definitions.ClientDo) error {
	// construct two clients to call out to keys server and
	// blockchain node.
	logger, err := loggerFromClientDo(do, "Send")
	if err != nil {
		return fmt.Errorf("Could not generate logging config from ClientDo: %s", err)
	}
	erisKeyClient := keys.NewErisKeyClient(do.SignAddrFlag, logger)
	erisNodeClient := client.NewErisNodeClient(do.NodeAddrFlag, logger)
	// form the send transaction
	sendTransaction, err := rpc.Send(erisNodeClient, erisKeyClient,
		do.PubkeyFlag, do.AddrFlag, do.ToFlag, do.AmtFlag, do.NonceFlag)
	if err != nil {
		fmt.Errorf("Failed on forming Send Transaction: %s", err)
	}
	// TODO: [ben] we carry over the sign bool, but always set it to true,
	// as we move away from and deprecate the api that allows sending unsigned
	// transactions and relying on (our) receiving node to sign it.
	txResult, err := rpc.SignAndBroadcast(do.ChainidFlag, erisNodeClient, erisKeyClient,
		sendTransaction, true, do.BroadcastFlag, do.WaitFlag)
	if err != nil {
		return fmt.Errorf("Failed on signing (and broadcasting) transaction: %s", err)
	}
	unpackSignAndBroadcast(txResult, logger)
	return nil
}
