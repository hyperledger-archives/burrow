package adapters

import (
	"github.com/eris-ltd/eris-db/logging/loggers"
	kitlog "github.com/go-kit/kit/log"
	"github.com/tendermint/log15"
)

type infoTraceLoggerAsLog15Handler struct {
	logger loggers.InfoTraceLogger
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

func InfoTraceLoggerAsLog15Handler(logger loggers.InfoTraceLogger) log15.Handler {
	return &infoTraceLoggerAsLog15Handler{
		logger: logger,
	}
}
