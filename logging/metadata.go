package logging

import (
	"time"

	"github.com/eris-ltd/eris-db/logging/loggers"
	"github.com/eris-ltd/eris-db/logging/structure"
	kitlog "github.com/go-kit/kit/log"
)

const (
	// To get the Caller information correct on the log, we need to count the
	// number of calls from a log call in the code to the time it hits a kitlog
	// context: [log call site (5), Info/Trace (4), MultipleChannelLogger.Log (3),
	// kitlog.Context.Log (2), kitlog.bindValues (1) (binding occurs),
	// kitlog.Caller (0), stack.caller]
	infoTraceLoggerCallDepth = 5
)

var defaultTimestampUTCValuer kitlog.Valuer = func() interface{} {
	return time.Now()
}

func WithMetadata(infoTraceLogger loggers.InfoTraceLogger) loggers.InfoTraceLogger {
	return infoTraceLogger.With(structure.TimeKey, defaultTimestampUTCValuer,
		structure.CallerKey, kitlog.Caller(infoTraceLoggerCallDepth))
}
