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

package lifecycle

// No package in ./logging/... should depend on lifecycle
import (
	"os"

	"time"

	"github.com/monax/eris-db/logging"
	"github.com/monax/eris-db/logging/adapters/stdlib"
	tmLog15adapter "github.com/monax/eris-db/logging/adapters/tendermint_log15"
	"github.com/monax/eris-db/logging/loggers"
	"github.com/monax/eris-db/logging/structure"

	kitlog "github.com/go-kit/kit/log"
	"github.com/streadway/simpleuuid"
	tmLog15 "github.com/tendermint/log15"
)

// Lifecycle provides a canonical source for eris loggers. Components should use the functions here
// to set up their root logger and capture any other logging output.

// Obtain a logger from a LoggingConfig
func NewLoggerFromLoggingConfig(LoggingConfig *logging.LoggingConfig) loggers.InfoTraceLogger {
	return NewStdErrLogger()
}

func NewStdErrLogger() loggers.InfoTraceLogger {
	logger := tmLog15adapter.Log15HandlerAsKitLogger(
		tmLog15.StreamHandler(os.Stderr, tmLog15.TerminalFormat()))
	return NewLogger(logger, logger)
}

// Provided a standard eris logger that outputs to the supplied underlying info and trace
// loggers
func NewLogger(infoLogger, traceLogger kitlog.Logger) loggers.InfoTraceLogger {
	infoTraceLogger := loggers.NewInfoTraceLogger(
		loggers.ErisFormatLogger(infoLogger),
		loggers.ErisFormatLogger(traceLogger))
	// Create a random ID based on start time
	uuid, _ := simpleuuid.NewTime(time.Now())
	var runId string
	if uuid != nil {
		runId = uuid.String()
	}
	return logging.WithMetadata(infoTraceLogger.With(structure.RunId, runId))
}

func CaptureTendermintLog15Output(infoTraceLogger loggers.InfoTraceLogger) {
	tmLog15.Root().SetHandler(
		tmLog15adapter.InfoTraceLoggerAsLog15Handler(infoTraceLogger.
			With(structure.CapturedLoggingSourceKey, "tendermint_log15")))
}

func CaptureStdlibLogOutput(infoTraceLogger loggers.InfoTraceLogger) {
	stdlib.CaptureRootLogger(infoTraceLogger.
		With(structure.CapturedLoggingSourceKey, "stdlib_log"))
}
