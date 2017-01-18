package logging

import (
	"github.com/eris-ltd/eris-db/logging/loggers"
	"github.com/eris-ltd/eris-db/logging/structure"
	"github.com/eris-ltd/eris-db/util/slice"
	kitlog "github.com/go-kit/kit/log"
)

// Helper functions for InfoTraceLoggers, sort of extension methods to loggers
// to centralise and establish logging conventions on top of in with the base
// logging interface

// Record structured Info log line with a message and conventional keys
func InfoMsgVals(logger loggers.InfoTraceLogger, message string, vals ...interface{}) {
	MsgVals(kitlog.LoggerFunc(logger.Info), message, vals...)
}

// Record structured Trace log line with a message and conventional keys
func TraceMsgVals(logger loggers.InfoTraceLogger, message string, vals ...interface{}) {
	MsgVals(kitlog.LoggerFunc(logger.Trace), message, vals...)
}

// Record structured Info log line with a message
func InfoMsg(logger loggers.InfoTraceLogger, message string, keyvals ...interface{}) {
	Msg(kitlog.LoggerFunc(logger.Info), message, keyvals...)
}

// Record structured Trace log line with a message
func TraceMsg(logger loggers.InfoTraceLogger, message string, keyvals ...interface{}) {
	Msg(kitlog.LoggerFunc(logger.Trace), message, keyvals...)
}

// Record a structured log line with a message
func Msg(logger kitlog.Logger, message string, keyvals ...interface{}) error {
	prepended := slice.CopyPrepend(keyvals, structure.MessageKey, message)
	return logger.Log(prepended...)
}

// Record a structured log line with a message and conventional keys
func MsgVals(logger kitlog.Logger, message string, vals ...interface{}) error {
	keyvals := make([]interface{}, len(vals)*2)
	for i := 0; i < len(vals); i++ {
		kv := i * 2
		keyvals[kv] = structure.KeyFromValue(vals[i])
		keyvals[kv+1] = vals[i]
	}
	return Msg(logger, message, keyvals)
}
