package lifecycle

// No package in ./logging/... should depend on lifecycle

import (
	"os"

	"github.com/eris-ltd/eris-db/logging"
	"github.com/eris-ltd/eris-db/logging/adapters/stdlib"
	tmLog15adapter "github.com/eris-ltd/eris-db/logging/adapters/tendermint_log15"
	"github.com/eris-ltd/eris-db/logging/loggers"
	"github.com/eris-ltd/eris-db/logging/structure"
	kitlog "github.com/go-kit/kit/log"
	tmLog15 "github.com/tendermint/log15"
	"github.com/streadway/simpleuuid"
	"time"
)

// Obtain a default eris-client logger from config
func NewClientLoggerFromConfig(LoggingConfig logging.LoggingConfig) loggers.InfoTraceLogger {
	infoLogger := kitlog.NewLogfmtLogger(os.Stderr)
	traceLogger := kitlog.NewLogfmtLogger(os.Stderr)
	return logging.WithMetadata(loggers.NewInfoTraceLogger(infoLogger, traceLogger))
}

// Obtain a default eris-server (eris-db serve ...) logger
func NewServerLoggerFromConfig(LoggingConfig logging.LoggingConfig) loggers.InfoTraceLogger {
	infoLogger := kitlog.NewLogfmtLogger(os.Stderr)
	traceLogger := kitlog.NewLogfmtLogger(os.Stderr)
	return logging.WithMetadata(loggers.NewInfoTraceLogger(infoLogger, traceLogger))
}

func NewStdErrLogger() loggers.InfoTraceLogger {
	logger := tmLog15adapter.Log15HandlerAsKitLogger(
		tmLog15.StreamHandler(os.Stderr, tmLog15.TerminalFormat()))
	return NewLogger(logger, logger)
}

func NewLogger(infoLogger, traceLogger kitlog.Logger) loggers.InfoTraceLogger {
	infoTraceLogger := loggers.NewInfoTraceLogger(infoLogger, traceLogger)
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
			With(structure.ComponentKey, "tendermint_log15")))
}

func CaptureStdlibLogOutput(infoTraceLogger loggers.InfoTraceLogger) {
	stdlib.CaptureRootLogger(infoTraceLogger.
		With(structure.ComponentKey, "stdlib_log"))
}
