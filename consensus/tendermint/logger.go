package tendermint

import (
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/tendermint/tmlibs/log"
)

type tendermintLogger struct {
	logger logging_types.InfoTraceLogger
}

func NewLogger(logger logging_types.InfoTraceLogger) *tendermintLogger {
	return &tendermintLogger{
		logger: logger,
	}
}

func (tml *tendermintLogger) Info(msg string, keyvals ...interface{}) {
	logging.InfoMsg(tml.logger, msg, keyvals...)
}

func (tml *tendermintLogger) Error(msg string, keyvals ...interface{}) {
	logging.InfoMsg(tml.logger, msg, keyvals...)
}

func (tml *tendermintLogger) Debug(msg string, keyvals ...interface{}) {
	logging.TraceMsg(tml.logger, msg, keyvals...)
}

func (tml *tendermintLogger) With(keyvals ...interface{}) log.Logger {
	return &tendermintLogger{
		logger: tml.logger.With(keyvals...),
	}
}
