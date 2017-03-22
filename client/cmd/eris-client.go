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

package commands

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eris-ltd/eris-db/definitions"
	"github.com/eris-ltd/eris-db/version"
)

// Global flags for persistent flags
var clientDo *definitions.ClientDo

var ErisClientCmd = &cobra.Command{
	Use:   "eris-client",
	Short: "Eris-client interacts with a running Eris chain.",
	Long: `Eris-client interacts with a running Eris chain.

Made with <3 by Monax Industries.

Complete documentation is available at https://monax.io/docs
` + "\nVERSION:\n " + version.VERSION,
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
	ErisClientCmd.AddCommand(buildTransactionCommand())
	ErisClientCmd.AddCommand(buildStatusCommand())

	buildGenesisGenCommand()
	ErisClientCmd.AddCommand(GenesisGenCmd)

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
