package loggers

import (
	"strings"

	kitlog "github.com/go-kit/kit/log"
)

// This represents an 'AND' type logger. When logged to it will log to each of
// the loggers in the slice.
type MultipleOutputLogger []kitlog.Logger

var _ kitlog.Logger = MultipleOutputLogger(nil)

func (mol MultipleOutputLogger) Log(keyvals ...interface{}) error {
	var errs []error
	for _, logger := range mol {
		err := logger.Log(keyvals...)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return combineErrors(errs)
}

// Creates a logger that forks log messages to each of its outputLoggers
func NewMultipleOutputLogger(outputLoggers ...kitlog.Logger) kitlog.Logger {
	return MultipleOutputLogger(outputLoggers)
}

type multipleErrors []error

func combineErrors(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return multipleErrors(errs)
	}
}

func (errs multipleErrors) Error() string {
	var errStrings []string
	for _, err := range errs {
		errStrings = append(errStrings, err.Error())
	}
	return strings.Join(errStrings, ";")
}
