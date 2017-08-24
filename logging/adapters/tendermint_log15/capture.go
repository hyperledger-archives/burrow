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

package tendermint_log15

import (
	kitlog "github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging/types"
	"github.com/tendermint/log15"
)

type infoTraceLoggerAsLog15Handler struct {
	logger types.InfoTraceLogger
}

var _ log15.Handler = (*infoTraceLoggerAsLog15Handler)(nil)

type log15HandlerAsKitLogger struct {
	handler log15.Handler
}

var _ kitlog.Logger = (*log15HandlerAsKitLogger)(nil)

func (l *log15HandlerAsKitLogger) Log(keyvals ...interface{}) error {
	record := LogLineToRecord(keyvals...)
	return l.handler.Log(record)
}

func (h *infoTraceLoggerAsLog15Handler) Log(record *log15.Record) error {
	if record.Lvl < log15.LvlDebug {
		// Send to Critical, Warning, Error, and Info to the Info channel
		h.logger.Info(RecordToLogLine(record)...)
	} else {
		// Send to Debug to the Trace channel
		h.logger.Trace(RecordToLogLine(record)...)
	}
	return nil
}

func Log15HandlerAsKitLogger(handler log15.Handler) kitlog.Logger {
	return &log15HandlerAsKitLogger{
		handler: handler,
	}
}

func InfoTraceLoggerAsLog15Handler(logger types.InfoTraceLogger) log15.Handler {
	return &infoTraceLoggerAsLog15Handler{
		logger: logger,
	}
}
