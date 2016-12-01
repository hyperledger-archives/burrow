package loggers

import (
	"fmt"

	kitlog "github.com/go-kit/kit/log"
)

// We use go-kit log's Context (a Logger with a context from which you can
// generate further nested contexts) as our main logger type. In order to have
// independently consumed channels of logs we maintain a map of contexts.
//
// This intended as a primitive to implement a routing logger or leveled logger
// where the channel names are compile-time constants of such a module.
type MultipleChannelLogger interface {
	// Log the message formed of keyvals to the the channelName logging channel
	Log(channelName string, keyvals ...interface{}) error

	// Establish a context by appending contextual key-values to any existing
	// contextual values
	With(keyvals ...interface{}) MultipleChannelLogger

	// Establish a context by prepending contextual key-values to any existing
	// contextual values
	WithPrefix(keyvals ...interface{}) MultipleChannelLogger
}

type multipleChannelLogger struct {
	loggers map[string]*kitlog.Context
}

// Pass in the map of named output loggers and get back a MultipleChannelLogger
// Each output logger will be wrapped as a ChannelLogger so will be non-blocking
// on calls to Log and provided the output loggers are not written to by any other
// means calls to MultipleChannelLogger will be safe for concurrent access
func NewMultipleChannelLogger(channelLoggers map[string]kitlog.Logger) MultipleChannelLogger {
	mcl := multipleChannelLogger{
		loggers: make(map[string]*kitlog.Context, len(channelLoggers)),
	}
	for name, logger := range channelLoggers {
		// Make a channel with a nil logger a noop
		if logger != nil {
			mcl.loggers[name] = kitlog.NewContext(NonBlockingLogger(logger)).
				With("channel", name)
		}
	}
	return &mcl
}

// Like go-kit log's Log method only logs a message to the specified channelName
// which must be a member of this MultipleChannelLogger
func (mcl *multipleChannelLogger) Log(channelName string, keyvals ...interface{}) error {
	logger := mcl.loggers[channelName]
	if logger == nil {
		return fmt.Errorf("Could not log to channel '%s', since it is not "+
			"registered with this MultipleChannelLogger (the underlying logger may "+
			"have been nil when passed to NewMultipleChannelLogger)", channelName)
	}
	return logger.Log(keyvals...)
}

// Like go-kit log's With method only it establishes a new Context for each
// logging channel
func (mcl *multipleChannelLogger) With(keyvals ...interface{}) MultipleChannelLogger {
	loggers := make(map[string]*kitlog.Context, len(mcl.loggers))
	for name, logger := range mcl.loggers {
		loggers[name] = logger.With(keyvals...)
	}
	return &multipleChannelLogger{
		loggers: loggers,
	}
}

// Like go-kit log's With method only it establishes a new Context for each
// logging channel
func (mcl *multipleChannelLogger) WithPrefix(keyvals ...interface{}) MultipleChannelLogger {
	loggers := make(map[string]*kitlog.Context, len(mcl.loggers))
	for name, logger := range mcl.loggers {
		loggers[name] = logger.WithPrefix(keyvals...)
	}
	return &multipleChannelLogger{
		loggers: loggers,
	}
}
