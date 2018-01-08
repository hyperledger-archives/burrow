package loggers

import kitlog "github.com/go-kit/kit/log"

// Filter logger allows us to filter lines logged to it before passing on to underlying
// output logger
// Creates a logger that removes lines from output when the predicate evaluates true
func FilterLogger(outputLogger kitlog.Logger, predicate func(keyvals []interface{}) bool) kitlog.Logger {
	return kitlog.LoggerFunc(func(keyvals ...interface{}) error {
		if !predicate(keyvals) {
			return outputLogger.Log(keyvals...)
		}
		return nil
	})
}
