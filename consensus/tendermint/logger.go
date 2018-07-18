package tendermint

import (
	"github.com/hyperledger/burrow/logging"
	"github.com/tendermint/tmlibs/log"
)

type tendermintLogger struct {
	logger *logging.Logger
}

func NewLogger(logger *logging.Logger) log.Logger {
	return &tendermintLogger{
		logger: logger,
	}
}

func (tml *tendermintLogger) Info(msg string, keyvals ...interface{}) {
	tml.logger.InfoMsg(msg, keyvals...)
}

func (tml *tendermintLogger) Error(msg string, keyvals ...interface{}) {
	tml.logger.InfoMsg(msg, keyvals...)
}

func (tml *tendermintLogger) Debug(msg string, keyvals ...interface{}) {
	tml.logger.TraceMsg(msg, keyvals...)
}

func (tml *tendermintLogger) With(keyvals ...interface{}) log.Logger {
	return &tendermintLogger{
		logger: tml.logger.With(keyvals...),
	}
}
