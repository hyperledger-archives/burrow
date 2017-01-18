// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package methods

import (
	log "github.com/eris-ltd/eris-logger"

	"github.com/eris-ltd/eris-db/client"
	"github.com/eris-ltd/eris-db/client/rpc"
	"github.com/eris-ltd/eris-db/definitions"
	"github.com/eris-ltd/eris-db/keys"
)

func Send(do *definitions.ClientDo) {
	// construct two clients to call out to keys server and
	// blockchain node.
	erisKeyClient := keys.NewErisKeyClient(do.SignAddrFlag)
	erisNodeClient := client.NewErisNodeClient(do.NodeAddrFlag)
	// form the send transaction
	sendTransaction, err := rpc.Send(erisNodeClient, erisKeyClient,
		do.PubkeyFlag, do.AddrFlag, do.ToFlag, do.AmtFlag, do.NonceFlag)
	if err != nil {
		log.Fatalf("Failed on forming Send Transaction: %s", err)
		return
	}
	// TODO: [ben] we carry over the sign bool, but always set it to true,
	// as we move away from and deprecate the api that allows sending unsigned
	// transactions and relying on (our) receiving node to sign it.
	unpackSignAndBroadcast(
		rpc.SignAndBroadcast(do.ChainidFlag, erisNodeClient,
			erisKeyClient, sendTransaction, true, do.BroadcastFlag, do.WaitFlag))
}
