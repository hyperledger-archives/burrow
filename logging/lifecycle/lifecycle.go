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

	"github.com/hyperledger/burrow/logging/adapters/stdlib"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/logging/structure"

	"fmt"

	"github.com/eapache/channels"
	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging"
)

// Lifecycle provides a canonical source for burrow loggers. Components should use the functions here
// to set up their root logger and capture any other logging output.

// Obtain a logger from a LoggingConfig
func NewLoggerFromLoggingConfig(loggingConfig *logconfig.LoggingConfig) (*logging.Logger, error) {
	if loggingConfig == nil {
		return NewStdErrLogger()
	} else {
		outputLogger, errCh, err := loggerFromLoggingConfig(loggingConfig)
		if err != nil {
			return nil, err
		}
		logger := logging.NewLogger(outputLogger)
		if loggingConfig.ExcludeTrace {
			logger.Trace = log.NewNopLogger()
		}
		go func() {
			err := <-errCh.Out()
			if err != nil {
				fmt.Printf("Logging error: %v", err)
			}
		}()
		return logger, nil
	}
}

// Hot swap logging config by replacing output loggers of passed InfoTraceLogger
// with those built from loggingConfig
func SwapOutputLoggersFromLoggingConfig(logger *logging.Logger, loggingConfig *logconfig.LoggingConfig) (error, channels.Channel) {
	outputLogger, errCh, err := loggerFromLoggingConfig(loggingConfig)
	if err != nil {
		return err, channels.NewDeadChannel()
	}
	logger.SwapOutput(outputLogger)
	return nil, errCh
}

func NewStdErrLogger() (*logging.Logger, error) {
	outputLogger, err := loggers.NewStreamLogger(os.Stderr, loggers.TerminalFormat)
	if err != nil {
		return nil, err
	}
	return logging.NewLogger(outputLogger), nil
}

func JustLogger(logger *logging.Logger, _ channels.Channel) *logging.Logger {
	return logger
}

func CaptureStdlibLogOutput(logger *logging.Logger) {
	stdlib.CaptureRootLogger(logger.With(structure.CapturedLoggingSourceKey, "stdlib_log"))
}

// Helpers
func loggerFromLoggingConfig(loggingConfig *logconfig.LoggingConfig) (log.Logger, channels.Channel, error) {
	outputLogger, _, err := loggingConfig.RootSink.BuildLogger()
	if err != nil {
		return nil, nil, err
	}
	var errCh channels.Channel = channels.NewDeadChannel()
	var logger log.Logger = loggers.BurrowFormatLogger(outputLogger)
	if loggingConfig.NonBlocking {
		logger, errCh = loggers.NonBlockingLogger(logger)
		return logger, errCh, nil
	}
	return logger, errCh, err
}
