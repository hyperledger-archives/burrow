package server

import (
	"fmt"
	"github.com/tendermint/log15"
	"os"
)

var rootHandler log15.Handler

func Init(config *ServerConfig) {

	consoleLogLevel := config.Logging.ConsoleLogLevel

	// stdout handler
	handlers := []log15.Handler{}
	stdoutHandler := log15.LvlFilterHandler(
		getLevel(consoleLogLevel),
		log15.StreamHandler(os.Stdout, log15.TerminalFormat()),
	)
	handlers = append(handlers, stdoutHandler)

	if config.Logging.LogFile != "" {
		fileLogLevel := config.Logging.FileLogLevel
		fh, err := log15.FileHandler(config.Logging.LogFile, log15.LogfmtFormat())
		if err != nil {
			fmt.Println("Error creating log file: " + err.Error())
			os.Exit(1)
		}
		fileHandler := log15.LvlFilterHandler(getLevel(fileLogLevel), fh)
		handlers = append(handlers, fileHandler)
	}
	
	rootHandler = log15.MultiHandler(handlers...)

	// By setting handlers on the root, we handle events from all loggers.
	log15.Root().SetHandler(rootHandler)
}

// See binary/log for an example of usage.
func RootHandler() log15.Handler {
	return rootHandler
}

func New(ctx ...interface{}) log15.Logger {
	return log15.Root().New(ctx...)
}

func getLevel(lvlString string) log15.Lvl {
	lvl, err := log15.LvlFromString(lvlString)
	if err != nil {
		fmt.Printf("Invalid log level %v: %v", lvlString, err)
		os.Exit(1)
	}
	return lvl
}
