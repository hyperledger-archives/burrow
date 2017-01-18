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
	"fmt"
	"os"

	log "github.com/eris-ltd/eris-logger"

	"github.com/eris-ltd/eris-db/client/rpc"
)

func unpackSignAndBroadcast(result *rpc.TxResult, err error) {
	if err != nil {
		log.Fatalf("Failed on signing (and broadcasting) transaction: %s", err)
		os.Exit(1)
	}
	if result == nil {
		// if we don't provide --sign or --broadcast
		return
	}
	printResult := log.Fields{
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
