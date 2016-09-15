// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package commands

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	log "github.com/eris-ltd/eris-logger"

	"github.com/eris-ltd/eris-db/definitions"
	"github.com/eris-ltd/eris-db/version"
)

// Global flags for persistent flags
var clientDo *definitions.ClientDo

var ErisClientCmd = &cobra.Command{
	Use:   "eris-client",
	Short: "Eris-client interacts with a running Eris chain.",
	Long: `Eris-client interacts with a running Eris chain.

Made with <3 by Eris Industries.

Complete documentation is available at https://docs.erisindustries.com
` + "\nVERSION:\n " + version.VERSION,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.WarnLevel)
		if clientDo.Verbose {
			log.SetLevel(log.InfoLevel)
		} else if clientDo.Debug {
			log.SetLevel(log.DebugLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func Execute() {
	InitErisClientInit()
	AddGlobalFlags()
	AddClientCommands()
	ErisClientCmd.Execute()
}

func InitErisClientInit() {
	// initialise an empty ClientDo struct for command execution
	clientDo = definitions.NewClientDo()
}

func AddGlobalFlags() {
	ErisClientCmd.PersistentFlags().BoolVarP(&clientDo.Verbose, "verbose", "v", defaultVerbose(), "verbose output; more output than no output flags; less output than debug level; default respects $ERIS_CLIENT_VERBOSE")
	ErisClientCmd.PersistentFlags().BoolVarP(&clientDo.Debug, "debug", "d", defaultDebug(), "debug level output; the most output available for eris-client; if it is too chatty use verbose flag; default respects $ERIS_CLIENT_DEBUG")
}

func AddClientCommands() {
	buildTransactionCommand()
	ErisClientCmd.AddCommand(TransactionCmd)
}

//------------------------------------------------------------------------------
// Defaults

// defaultVerbose is set to false unless the ERIS_CLIENT_VERBOSE environment
// variable is set to a parsable boolean.
func defaultVerbose() bool {
	return setDefaultBool("ERIS_CLIENT_VERBOSE", false)
}

// defaultDebug is set to false unless the ERIS_CLIENT_DEBUG environment
// variable is set to a parsable boolean.
func defaultDebug() bool {
	return setDefaultBool("ERIS_CLIENT_DEBUG", false)
}

// setDefaultBool returns the provided default value if the environment variable
// is not set or not parsable as a bool.
func setDefaultBool(environmentVariable string, defaultValue bool) bool {
	value := os.Getenv(environmentVariable)
	if value != "" {
		if parsedValue, err := strconv.ParseBool(value); err == nil {
			return parsedValue
		}
	}
	return defaultValue
}

func setDefaultString(envVar, def string) string {
	env := os.Getenv(envVar)
	if env != "" {
		return env
	}
	return def
}

func setDefaultStringSlice(envVar string, def []string) []string {
	env := os.Getenv(envVar)
	if env != "" {
		return strings.Split(env, ",")
	}
	return def
}
