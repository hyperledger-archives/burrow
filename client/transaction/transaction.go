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

package transaction

import (
	"fmt"
	// "io/ioutil"
	"os"

	log "github.com/eris-ltd/eris-logger"

	"github.com/eris-ltd/eris-db/client/core"
	"github.com/eris-ltd/eris-db/definitions"
)

func Send(do *definitions.ClientDo) {
	// form the send transaction
	sendTransaction, err := core.Send(do.NodeAddrFlag, do.SignAddrFlag,
		do.PubkeyFlag, do.AddrFlag, do.ToFlag, do.AmtFlag, do.NonceFlag)
	if err != nil {
		log.Fatalf("Failed on forming Send Transaction: %s", err)
		return
	}
	// TODO: [ben] we carry over the sign bool, but always set it to true,
	// as we move away from and deprecate the api that allows sending unsigned
	// transactions and relying on (our) receiving node to sign it. 
	unpackSignAndBroadcast(
		core.SignAndBroadcast(do.ChainidFlag, do.NodeAddrFlag,
		do.SignAddrFlag, sendTransaction, true, do.BroadcastFlag, do.WaitFlag))
}

func Call(do *definitions.ClientDo) {
	// form the call transaction
	callTransaction, err := core.Call(do.NodeAddrFlag, do.SignAddrFlag,
		do.PubkeyFlag, do.AddrFlag, do.ToFlag, do.AmtFlag, do.NonceFlag,
		do.GasFlag, do.FeeFlag, do.DataFlag)
	if err != nil {
		log.Fatalf("Failed on forming Call Transaction: %s", err)
		return
	}
	// TODO: [ben] we carry over the sign bool, but always set it to true,
	// as we move away from and deprecate the api that allows sending unsigned
	// transactions and relying on (our) receiving node to sign it. 
	unpackSignAndBroadcast(
		core.SignAndBroadcast(do.ChainidFlag, do.NodeAddrFlag,
		do.SignAddrFlag, callTransaction, true, do.BroadcastFlag, do.WaitFlag))
}

//----------------------------------------------------------------------
// Helper functions

func unpackSignAndBroadcast(result *core.TxResult, err error) {
	if err != nil {
		log.Fatalf("Failed on signing (and broadcasting) transaction: %s", err)
		os.Exit(1)
	}
	if result == nil {
		// if we don't provide --sign or --broadcast
		return
	}
	printResult := log.Fields {
		"transaction hash": fmt.Sprintf("%X", result.Hash),
	}
	if result.Address != nil {
		printResult["Contract Address"] = fmt.Sprintf("%X", result.Address)
	}
	if result.Return != nil {
		printResult["Block Hash"] = fmt.Sprintf("%X", result.BlockHash)
		printResult["Return Value"] = fmt.Sprintf("%X", result.Return)
		printResult["Exception"] = fmt.Sprintf("%s", result.Exception)
	}
	log.WithFields(printResult).Warn("Result")
}
