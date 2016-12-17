package loggers

import (
	"github.com/eris-ltd/eris-db/logging/structure"
	kitlog "github.com/go-kit/kit/log"
)

// Treat duplicate keys
type vectorValuedLogger struct {
	logger kitlog.Logger
}

var _ kitlog.Logger = &vectorValuedLogger{}

func (vvl *vectorValuedLogger) Log(keyvals ...interface{}) error {
	keys, vals := structure.KeyValuesVector(keyvals)
	kvs := make([]interface{}, len(keys)*2)
	for i := 0; i < len(keys); i++ {
		kv := i * 2
		key := keys[i]
		kvs[kv] = key
		kvs[kv+1] = vals[key]
	}
	return vvl.logger.Log(kvs...)
}

func VectorValuedLogger(logger kitlog.Logger) *vectorValuedLogger {
	return &vectorValuedLogger{logger: logger}
}
