package loggers

import (
	"fmt"
	"io"
	"os"
	"text/template"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/term"
	"github.com/hyperledger/burrow/logging/structure"
)

const (
	JSONFormat        = "json"
	LogfmtFormat      = "logfmt"
	TerminalFormat    = "terminal"
	defaultFormatName = TerminalFormat
)

const (
	newline = '\n'
)

func NewStreamLogger(writer io.Writer, format string) (kitlog.Logger, error) {
	var logger kitlog.Logger
	var err error
	switch format {
	case "":
		return NewStreamLogger(writer, defaultFormatName)
	case JSONFormat:
		logger = kitlog.NewJSONLogger(writer)
	case LogfmtFormat:
		logger = kitlog.NewLogfmtLogger(writer)
	case TerminalFormat:
		logger = term.NewLogger(writer, kitlog.NewLogfmtLogger, func(keyvals ...interface{}) term.FgBgColor {
			switch structure.Value(keyvals, structure.ChannelKey) {
			case structure.TraceChannelName:
				return term.FgBgColor{Fg: term.DarkGreen}
			default:
				return term.FgBgColor{Fg: term.Yellow}
			}
		})
	default:
		logger, err = NewTemplateLogger(writer, format, []byte{})
		if err != nil {
			return nil, fmt.Errorf("did not recognise format '%s' as named format and could not parse as "+
				"template: %v", format, err)
		}
	}
	// Don't log signals
	return kitlog.LoggerFunc(func(keyvals ...interface{}) error {
		if structure.Signal(keyvals) != "" {
			return nil
		}
		return logger.Log(keyvals...)
	}), nil
}

func NewFileLogger(path string, formatName string) (kitlog.Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	streamLogger, err := NewStreamLogger(f, formatName)
	if err != nil {
		return nil, err
	}
	return kitlog.LoggerFunc(func(keyvals ...interface{}) error {
		if structure.Signal(keyvals) == structure.SyncSignal {
			return f.Sync()
		}
		return streamLogger.Log(keyvals...)
	}), nil
}

func NewTemplateLogger(writer io.Writer, textTemplate string, recordSeparator []byte) (kitlog.Logger, error) {
	tmpl, err := template.New("template-logger").Parse(textTemplate)
	if err != nil {
		return nil, err
	}
	return kitlog.LoggerFunc(func(keyvals ...interface{}) error {
		err := tmpl.Execute(writer, structure.KeyValuesMap(keyvals))
		if err == nil {
			_, err = writer.Write(recordSeparator)
		}
		return err
	}), nil

}
