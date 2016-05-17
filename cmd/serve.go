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
  // "os"

  cobra "github.com/spf13/cobra"

  // common "github.com/eris-ltd/common/go/common"

  // definitions "github.com/eris-ltd/eris-db/definitions"
)

var ServeCmd = &cobra.Command {
  Use:   "serve",
  Short: "Eris-DB serve starts an eris-db node with client API enabled by default.",
  Long:  `Eris-DB serve starts an eris-db node with client API enabled by default.
The Eris-DB node is modularly configured for the consensus engine and application
manager.  The client API can be disabled.`,
  Example: `$ eris-db serve -- will start the Eris-DB node based on the configuration file in the current working directory,
$ eris-db serve myChainId --work-dir=/path/to/config -- will start the Eris-DB node based on the configuration file provided and assert the chain id matches.`,
  PreRun: func(cmd *cobra.Command, args []string) {
    // TODO: [ben] log marmotty welcome
  },
  Run: func(cmd *cobra.Command, args []string) {
    serve()
  },
}

// build the serve subcommand
func buildServeCommand() {
  addServeFlags()
}

func addServeFlags() {
  fmt.Println("Adding Serve flags")
  ServeCmd.PersistentFlags().StringVarP(&do.WorkDir, "work-dir", "w",
    defaultWorkDir(), "specify the working directory for the chain to run.  If omitted, and no path set in $ERIS_DB_WORKDIR, the current working directory is taken.")
}


//------------------------------------------------------------------------------
// functions

func serve() {
  //
  // load config from
  loadConfig()
  fmt.Printf("Served from %s \n", do.WorkDir)
}

//------------------------------------------------------------------------------
// Viper configuration

func loadConfig(conf *viper.Viper, path string) {
  conf.
}

//------------------------------------------------------------------------------
// Defaults

func defaultWorkDir() string {
  // if ERIS_DB_WORKDIR environment variable is not set, keep do.WorkDir empty
  return setDefaultString("ERIS_DB_WORKDIR", "")
}
