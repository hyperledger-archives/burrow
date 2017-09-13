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

	"github.com/hyperledger/burrow/logging/adapters/stdlib"
	"github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/logging/structure"

	"fmt"

	"github.com/eapache/channels"
	kitlog "github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/types"
	"github.com/streadway/simpleuuid"
)

// Lifecycle provides a canonical source for burrow loggers. Components should use the functions here
// to set up their root logger and capture any other logging output.

// Obtain a logger from a LoggingConfig
func NewLoggerFromLoggingConfig(loggingConfig *config.LoggingConfig) (types.InfoTraceLogger, error) {
	var logger types.InfoTraceLogger
	var errCh channels.Channel
	if loggingConfig == nil {
		logger, errCh = NewStdErrLogger()
	} else {
		outputLogger, err := infoTraceLoggerFromLoggingConfig(loggingConfig)
		if err != nil {
			return nil, err
		}
		logger, errCh = NewLogger(outputLogger)
	}
	go func() {
		err := <-errCh.Out()
		if err != nil {
			logger.Info("logging_error", err,
				"logging_config", loggingConfig.RootTOMLString(),
				"logger", fmt.Sprintf("%#v", logger))
		}
	}()

	return logger, nil
}

// Hot swap logging config by replacing output loggers of passed InfoTraceLogger
// with those built from loggingConfig
func SwapOutputLoggersFromLoggingConfig(logger types.InfoTraceLogger,
	loggingConfig *config.LoggingConfig) error {
	outputLogger, err := infoTraceLoggerFromLoggingConfig(loggingConfig)
	if err != nil {
		return err
	}
	logger.SwapOutput(outputLogger)
	return nil
}

func NewStdErrLogger() (types.InfoTraceLogger, channels.Channel) {
	logger := loggers.NewStreamLogger(os.Stderr, "terminal")
	return NewLogger(logger)
}

// Provided a standard logger that outputs to the supplied underlying outputLogger
func NewLogger(outputLogger kitlog.Logger) (types.InfoTraceLogger, channels.Channel) {
	infoTraceLogger, errCh := loggers.NewInfoTraceLogger(outputLogger)
	// Create a random ID based on start time
	uuid, _ := simpleuuid.NewTime(time.Now())
	var runId string
	if uuid != nil {
		runId = uuid.String()
	}
	return logging.WithMetadata(infoTraceLogger.With(structure.RunId, runId)), errCh
}

func CaptureStdlibLogOutput(infoTraceLogger types.InfoTraceLogger) {
	stdlib.CaptureRootLogger(infoTraceLogger.
		With(structure.CapturedLoggingSourceKey, "stdlib_log"))
}

// Helpers
func infoTraceLoggerFromLoggingConfig(loggingConfig *config.LoggingConfig) (kitlog.Logger, error) {
	outputLogger, _, err := loggingConfig.RootSink.BuildLogger()
	if err != nil {
		return nil, err
	}
	return outputLogger, nil
}
