package lifecycle

// No package in ./logging/... should depend on lifecycle

import (
	"os"

	"github.com/eris-ltd/eris-db/logging"
	tmLog15adapter "github.com/eris-ltd/eris-db/logging/adapters/tendermint_log15"
	"github.com/eris-ltd/eris-db/logging/loggers"
	"github.com/eris-ltd/eris-db/logging/structure"
	kitlog "github.com/go-kit/kit/log"
	tmLog15 "github.com/tendermint/log15"
)

func NewLoggerFromConfig(LoggingConfig logging.LoggingConfig) loggers.InfoTraceLogger {
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
	return logging.WithMetadata(infoTraceLogger)
}

func CaptureTendermintLog15Output(infoTraceLogger loggers.InfoTraceLogger) {
	tmLog15.Root().SetHandler(
		tmLog15adapter.InfoTraceLoggerAsLog15Handler(infoTraceLogger.
			With(structure.ComponentKey, "tendermint")))
}
