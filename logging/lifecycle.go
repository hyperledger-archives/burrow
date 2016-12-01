package logging

import (
	"os"

	"github.com/eris-ltd/eris-db/logging/loggers"
	kitlog "github.com/go-kit/kit/log"
)

const (
	MessageKey = "message"
	// To get the Caller information correct on the log, we need to count the
	// number of calls from a log call in the code to the time it hits a kitlog
	// context: [log call site (5), Info/Trace (4), MultipleChannelLogger.Log (3),
	// kitlog.Context.Log (2), kitlog.bindValues (1) (binding occurs),
	// kitlog.Caller (0), stack.caller]
	infoTraceLoggerCallDepth = 5
)

func NewLogger(LoggingConfig LoggingConfig) loggers.InfoTraceLogger {
	infoLogger := kitlog.NewLogfmtLogger(os.Stderr)
	traceLogger := kitlog.NewLogfmtLogger(os.Stderr)
	return loggers.NewInfoTraceLogger(infoLogger, traceLogger).
		With("timestamp_utc", kitlog.DefaultTimestampUTC,
			"caller", kitlog.Caller(infoTraceLoggerCallDepth))
}
