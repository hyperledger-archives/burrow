package loggers

import (
	"github.com/eris-ltd/eris-db/logging/structure"
	kitlog "github.com/go-kit/kit/log"
)

// Treat duplicate key-values as consecutive entries in a vector-valued lookup
type vectorValuedLogger struct {
	logger kitlog.Logger
}

var _ kitlog.Logger = &vectorValuedLogger{}

func (vvl *vectorValuedLogger) Log(keyvals ...interface{}) error {
	return vvl.logger.Log(structure.Vectorise(keyvals)...)
}

func VectorValuedLogger(logger kitlog.Logger) *vectorValuedLogger {
	return &vectorValuedLogger{logger: logger}
}
