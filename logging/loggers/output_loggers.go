package loggers

import (
	"io"

	"os"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/term"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/logging/types"
)

const (
	JSONFormat        = "json"
	LogfmtFormat      = "logfmt"
	TerminalFormat    = "terminal"
	defaultFormatName = TerminalFormat
)

func NewStreamLogger(writer io.Writer, formatName string) kitlog.Logger {
	var logger kitlog.Logger
	switch formatName {
	case JSONFormat:
		logger = kitlog.NewJSONLogger(writer)
	case LogfmtFormat:
		logger = kitlog.NewLogfmtLogger(writer)
	default:
		logger = term.NewLogger(writer, kitlog.NewLogfmtLogger, func(keyvals ...interface{}) term.FgBgColor {
			switch structure.Value(keyvals, structure.ChannelKey) {
			case types.TraceChannelName:
				return term.FgBgColor{Fg: term.DarkGreen}
			default:
				return term.FgBgColor{Fg: term.Yellow}
			}
		})
	}
	// Don't log signals
	return kitlog.LoggerFunc(func(keyvals ...interface{}) error {
		if logging.Signal(keyvals) != "" {
			return nil
		}
		return logger.Log(keyvals...)
	})
}

func NewFileLogger(path string, formatName string) (kitlog.Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	streamLogger := NewStreamLogger(f, formatName)
	return kitlog.LoggerFunc(func(keyvals ...interface{}) error {
		if logging.Signal(keyvals) == structure.SyncSignal {
			return f.Sync()
		}
		return streamLogger.Log(keyvals...)
	}), nil
}
