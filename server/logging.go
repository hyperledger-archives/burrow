// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"fmt"
	"os"

	"github.com/tendermint/log15"
)

var rootHandler log15.Handler

// This is basically the same code as in tendermint. Initialize root
// and maybe later also track the loggers here.
func InitLogger(config *ServerConfig) {

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
