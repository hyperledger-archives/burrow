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
	"github.com/monax/eris-db/client/rpc"
	"github.com/monax/eris-db/core"
	"github.com/monax/eris-db/definitions"
	"github.com/monax/eris-db/logging"
	"github.com/monax/eris-db/logging/lifecycle"
	"github.com/monax/eris-db/logging/loggers"
)

func unpackSignAndBroadcast(result *rpc.TxResult, logger loggers.InfoTraceLogger) {
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
