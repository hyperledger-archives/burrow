package lifecycle

// No package in ./logging/... should depend on lifecycle
import (
	"os"

	"time"

	"github.com/eris-ltd/eris-db/logging"
	"github.com/eris-ltd/eris-db/logging/adapters/stdlib"
	tmLog15adapter "github.com/eris-ltd/eris-db/logging/adapters/tendermint_log15"
	"github.com/eris-ltd/eris-db/logging/loggers"
	"github.com/eris-ltd/eris-db/logging/structure"

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
