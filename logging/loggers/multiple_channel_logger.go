package loggers

import (
	"fmt"

	"github.com/eris-ltd/eris-db/logging/structure"
	kitlog "github.com/go-kit/kit/log"
)

// This represents a 'SELECT ONE' type logger. When logged to it will search
// for the ChannelKey field, look that up in its map and send the log line there
// Otherwise logging is a noop (but an error will be returned - which is optional)
type MultipleChannelLogger map[string]kitlog.Logger

var _ kitlog.Logger = MultipleChannelLogger(nil)

// Like go-kit log's Log method only logs a message to the specified channelName
// which must be a member of this MultipleChannelLogger
func (mcl MultipleChannelLogger) Log(keyvals ...interface{}) error {
	channel := structure.Value(keyvals, structure.ChannelKey)
	if channel == nil {
		return fmt.Errorf("MultipleChannelLogger could not select channel because"+
			" '%s' was not set in log message", structure.ChannelKey)
	}
	channelName, ok := channel.(string)
	if !ok {
		return fmt.Errorf("MultipleChannelLogger could not select channel because"+
			" channel was set to non-string value %v", channel)
	}
	logger := mcl[channelName]
	if logger == nil {
		return fmt.Errorf("Could not log to channel '%s', since it is not "+
			"registered with this MultipleChannelLogger (the underlying logger may "+
			"have been nil when passed to NewMultipleChannelLogger)", channelName)
	}
	return logger.Log(keyvals...)
}
