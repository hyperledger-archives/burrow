package stdlib

import (
	"io"
	"log"

	"github.com/eris-ltd/eris-db/logging/loggers"
	kitlog "github.com/go-kit/kit/log"
)

func Capture(stdLibLogger log.Logger,
	logger loggers.InfoTraceLogger) io.Writer {
	adapter := newAdapter(logger)
	stdLibLogger.SetOutput(adapter)
	return adapter
}

func CaptureRootLogger(logger loggers.InfoTraceLogger) io.Writer {
	adapter := newAdapter(logger)
	log.SetOutput(adapter)
	return adapter
}

func newAdapter(logger loggers.InfoTraceLogger) io.Writer {
	return kitlog.NewStdlibAdapter(logger)
}
