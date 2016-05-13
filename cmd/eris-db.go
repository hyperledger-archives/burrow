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
  "fmt"
  "os"

  cobra "github.com/spf13/cobra"

  common "github.com/eris-ltd/common/go/common"

  defintion "github.com/eris-ltd/eris-db/defintions"
  version   "github.com/eris-ltd/eris-db/version"
)

const VERSION = version.VERSION

// Global Do struct
var do *definitions.Do

var ErisDbCmd = &cobra.Command {
  Use:   "eris-db"
  Short: "Eris-DB is the beating heart of the eris chain."
  Long:  `Eris-DB is the beating heart of the eris chain.  Eris-DB combines
a modular consensus engine and application manager to run a chain to suit
your needs.

Made with <3 by Eris Industries.

Complete documentation is available at https://docs.erisindustries.com
` + "\nVERSION:\n " + VERSION,
  PersistentPreRun: func(cmd *cobra.Command, args [string]) {
    // TODO: [ben] set up eris logger after glide resolution of logrus
  },
  Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func Execute() {
  InitErisDb()
  AddGlobalFlags()
  AddCommands()
  ErisDbCmd.Execute()
}

func InitErisDb() {
  // initialise an empty do struct for command execution
	do = definitions.NowDo()
}

func AddCommands() {
  buildServeCommand()
  ErisDbCmd.AddCommand()
}

func AddGlobalFlags() {
	ErisDbCmd.PersistentFlags().BoolVarP(&do.Verbose, "verbose", "v", defaultVerbose(), "verbose output; more output than no output flags; less output than debug level; default respects $ERIS_DB_VERBOSE")
	ErisDbCmd.PersistentFlags().BoolVarP(&do.Debug, "debug", "d", defaultDebug(), "debug level output; the most output available for eris-db; if it is too chatty use verbose flag; default respects $ERIS_DB_DEBUG")
	ErisDbCmd.PersistentFlags().BoolVarP(&do.Output, "output", "o", defaultOutput(), "should eris-db provide an output of its execution; default respects $ERIS_DB_OUTPUT")
}

//------------------------------------------------------------------------------
// Defaults

// defaultVerbose is set to false unless the ERIS_DB_VERBOSE environment
// variable is set to a parsable boolean.
func defaultVerbose() bool {
  return setDefaultBool("ERIS_DB_VERBOSE", false)
}

// defaultDebug is set to false unless the ERIS_DB_DEBUG environment
// variable is set to a parsable boolean.
func defaultDebug() bool {
  return setDefaultBool("ERIS_DB_DEBUG", false)
}

// defaultOutput is set to true unless the ERIS_DB_OUTPUT environment
// variable is set to a parsable boolean.
func defaultOutput() bool {
  return setDefaultBool("ERIS_DB_OUTPUT", true)
}

// setDefaultBool returns the provided default value if the environment variab;e
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
