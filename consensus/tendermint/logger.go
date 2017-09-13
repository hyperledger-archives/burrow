package tendermint

import (
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/tendermint/tmlibs/log"
)

type tmLogger struct {
	logger logging_types.InfoTraceLogger
}

func NewLogger(logger logging_types.InfoTraceLogger) *tmLogger {
	return &tmLogger{
		logger: logger,
	}
}

func (tml *tmLogger) Info(msg string, keyvals ...interface{}) error {
	return logging.InfoMsg(tml.logger, msg, keyvals...)
}

func (tml *tmLogger) Error(msg string, keyvals ...interface{}) error {
	return logging.InfoMsg(tml.logger, msg, keyvals...)
}

func (tml *tmLogger) Debug(msg string, keyvals ...interface{}) error {
	return logging.TraceMsg(tml.logger, msg, keyvals...)
}

func (tml *tmLogger) With(keyvals ...interface{}) log.Logger {
	return &tmLogger{
		logger: tml.logger.With(keyvals...),
	}
}
