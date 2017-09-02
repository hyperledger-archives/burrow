package loggers

import (
	"io"

	"log/syslog"
	"net/url"

	kitlog "github.com/go-kit/kit/log"
	log15a "github.com/hyperledger/burrow/logging/adapters/tendermint_log15"
	"github.com/tendermint/log15"
)

const (
	syslogPriority    = syslog.LOG_LOCAL0
	JSONFormat        = "json"
	LogfmtFormat      = "logfmt"
	TerminalFormat    = "terminal"
	defaultFormatName = TerminalFormat
)

func NewStreamLogger(writer io.Writer, formatName string) kitlog.Logger {
	switch formatName {
	case JSONFormat:
		return kitlog.NewJSONLogger(writer)
	case LogfmtFormat:
		return kitlog.NewLogfmtLogger(writer)
	default:
		return log15a.Log15HandlerAsKitLogger(log15.StreamHandler(writer,
			format(formatName)))
	}
}

func NewFileLogger(path string, formatName string) (kitlog.Logger, error) {
	handler, err := log15.FileHandler(path, format(formatName))
	return log15a.Log15HandlerAsKitLogger(handler), err
}

func NewRemoteSyslogLogger(url *url.URL, tag, formatName string) (kitlog.Logger, error) {
	handler, err := log15.SyslogNetHandler(url.Scheme, url.Host, syslogPriority,
		tag, format(formatName))
	if err != nil {
		return nil, err
	}
	return log15a.Log15HandlerAsKitLogger(handler), nil
}

func NewSyslogLogger(tag, formatName string) (kitlog.Logger, error) {
	handler, err := log15.SyslogHandler(syslogPriority, tag, format(formatName))
	if err != nil {
		return nil, err
	}
	return log15a.Log15HandlerAsKitLogger(handler), nil
}

func format(name string) log15.Format {
	switch name {
	case JSONFormat:
		return log15.JsonFormat()
	case LogfmtFormat:
		return log15.LogfmtFormat()
	case TerminalFormat:
		return log15.TerminalFormat()
	default:
		return format(defaultFormatName)
	}
}
