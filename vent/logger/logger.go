package logger

import (
	"os"

	kitlog "github.com/go-kit/kit/log"
	kitlevel "github.com/go-kit/kit/log/level"
)

// Logger wraps a go-kit logger
type Logger struct {
	Log kitlog.Logger
}

// NewLogger creates a new logger based on the given level
func NewLogger(level string) *Logger {
	log := kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stdout))
	switch level {
	case "error":
		log = kitlevel.NewFilter(log, kitlevel.AllowError()) // only error logs
	case "warn":
		log = kitlevel.NewFilter(log, kitlevel.AllowWarn()) // warn + error logs
	case "info":
		log = kitlevel.NewFilter(log, kitlevel.AllowInfo()) // info + warn + error logs
	case "debug":
		log = kitlevel.NewFilter(log, kitlevel.AllowDebug()) // all logs
	default:
		log = kitlevel.NewFilter(log, kitlevel.AllowNone()) // no logs
	}

	log = kitlog.With(log, "service", "vent")
	log = kitlog.With(log, "ts", kitlog.DefaultTimestampUTC)
	log = kitlog.With(log, "caller", kitlog.Caller(4))

	return &Logger{
		Log: log,
	}
}

// NewLoggerFromKitlog creates a logger from a go-kit logger
func NewLoggerFromKitlog(log kitlog.Logger) *Logger {
	return &Logger{
		Log: log,
	}
}

// Error prints an error log
func (l *Logger) Error(args ...interface{}) {
	kitlevel.Error(l.Log).Log(args...)
}

// Warn prints a warning log
func (l *Logger) Warn(args ...interface{}) {
	kitlevel.Warn(l.Log).Log(args...)
}

// Info prints an information log
func (l *Logger) Info(args ...interface{}) {
	kitlevel.Info(l.Log).Log(args...)
}

// Debug prints a debug log
func (l *Logger) Debug(args ...interface{}) {
	kitlevel.Debug(l.Log).Log(args...)
}
