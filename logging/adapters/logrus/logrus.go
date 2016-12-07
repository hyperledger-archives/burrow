package adapters

import (
	"github.com/Sirupsen/logrus"
	kitlog "github.com/go-kit/kit/log"
)

type logrusLogger struct {
	logger logrus.Logger
}

var _ kitlog.Logger = (*logrusLogger)(nil)

func NewLogrusLogger(logger logrus.Logger) *logrusLogger {
	return &logrusLogger{
		logger: logger,
	}
}

func (ll *logrusLogger) Log(keyvals... interface{}) error {
	return nil
}

