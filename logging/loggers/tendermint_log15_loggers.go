package loggers

import (
	"io"

	kitlog "github.com/go-kit/kit/log"
	"github.com/tendermint/log15"
	log15a "github.com/eris-ltd/eris-db/logging/adapters/tendermint_log15"
)

func NewStreamLogger(writer io.Writer) kitlog.Logger {
	return log15a.Log15HandlerAsKitLogger(log15.StreamHandler(writer, log15.TerminalFormat()))
}

func NewFileLogger(path string) (kitlog.Logger, error) {
	handler, err := log15.FileHandler(path, log15.LogfmtFormat())
	return log15a.Log15HandlerAsKitLogger(handler), err
}
