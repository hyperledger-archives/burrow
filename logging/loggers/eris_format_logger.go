package loggers

import (
	"fmt"

	"github.com/eris-ltd/eris-db/logging/structure"
	kitlog "github.com/go-kit/kit/log"
)

// Logger that implements some formatting conventions for eris-db and eris-client
// This is intended for applying consistent value formatting before the final 'output' logger;
// we should avoid prematurely formatting values here if it is useful to let the output logger
// decide how it wants to display values. Ideal candidates for 'early' formatting here are types that
// we control and generic output loggers are unlikely to know about.
type erisFormatLogger struct {
	logger kitlog.Logger
}

var _ kitlog.Logger = &erisFormatLogger{}

func (efl *erisFormatLogger) Log(keyvals ...interface{}) error {
	return efl.logger.Log(structure.MapKeyValues(keyvals, erisFormatKeyValueMapper)...)
}

func erisFormatKeyValueMapper(key, value interface{}) (interface{}, interface{}) {
	switch key {
	default:
		switch v := value.(type) {
		case []byte:
			return key, fmt.Sprintf("%X", v)
		}
	}
	return key, value
}

func ErisFormatLogger(logger kitlog.Logger) *erisFormatLogger {
	return &erisFormatLogger{logger: logger}
}
