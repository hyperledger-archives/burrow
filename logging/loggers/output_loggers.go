package loggers

import (
	"io"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/term"
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
	switch formatName {
	case JSONFormat:
		return kitlog.NewJSONLogger(writer)
	case LogfmtFormat:
		return kitlog.NewLogfmtLogger(writer)
	default:
		return term.NewLogger(writer, kitlog.NewLogfmtLogger, func(keyvals ...interface{}) term.FgBgColor {
			switch structure.Value(keyvals, structure.ChannelKey) {
			case types.TraceChannelName:
				return term.FgBgColor{Fg: term.DarkGray}
			default:
				return term.FgBgColor{}
			}
		})
	}
}
