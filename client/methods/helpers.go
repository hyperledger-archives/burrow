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
	"github.com/eris-ltd/eris-db/client/rpc"
	"github.com/eris-ltd/eris-db/core"
	"github.com/eris-ltd/eris-db/definitions"
	"github.com/eris-ltd/eris-db/logging"
	"github.com/eris-ltd/eris-db/logging/lifecycle"
	"github.com/eris-ltd/eris-db/logging/loggers"
	"github.com/eris-ltd/eris-db/util"
)

func unpackSignAndBroadcast(result *rpc.TxResult, err error, logger loggers.InfoTraceLogger) {
	if err != nil {
		util.Fatalf("Failed on signing (and broadcasting) transaction: %s", err)
	}
	if result == nil {
		// if we don't provide --sign or --broadcast
		return
	}

	logger = logger.With("transaction hash", result.Hash)

	if result.Address != nil {
		logger = logger.With("Contract Address", result.Address)
	}

	if result.Return != nil {
		logger = logger.With("Block Hash", result.BlockHash,
			"Return Value", result.Return,
			"Exception", result.Exception,
		)
	}

	logging.InfoMsg(logger, "SignAndBroadcast result")
}

func loggerFromClientDo(do *definitions.ClientDo, scope string) (loggers.InfoTraceLogger, error) {
	lc, err := core.LoadLoggingConfigFromClientDo(do)
	if err != nil {
		return nil, err
	}
	return logging.WithScope(lifecycle.NewLoggerFromLoggingConfig(lc), scope), nil
}
