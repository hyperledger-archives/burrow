package loggers

import kitlog "github.com/go-kit/kit/log"

// Filter logger allows us to filter lines logged to it before passing on to underlying
// output logger
type filterLogger struct {
	logger    kitlog.Logger
	predicate func(keyvals []interface{}) bool
}

var _ kitlog.Logger = (*filterLogger)(nil)

func (fl filterLogger) Log(keyvals ...interface{}) error {
	if !fl.predicate(keyvals) {
		return fl.logger.Log(keyvals...)
	}
	return nil
}

// Creates a logger that removes lines from output when the predicate evaluates true
func NewFilterLogger(outputLogger kitlog.Logger,
	predicate func(keyvals []interface{}) bool) kitlog.Logger {
	return &filterLogger{
		logger:    outputLogger,
		predicate: predicate,
	}
}
