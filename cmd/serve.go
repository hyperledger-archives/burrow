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

  cobra "github.com/spf13/cobra"

  log "github.com/eris-ltd/eris-logger"
)

var ServeCmd = &cobra.Command {
  Use:   "serve",
  Short: "Eris-DB serve starts an eris-db node with client API enabled by default.",
  Long:  `Eris-DB serve starts an eris-db node with client API enabled by default.
The Eris-DB node is modularly configured for the consensus engine and application
manager.  The client API can be disabled.`,
  Example: `$ eris-db serve -- will start the Eris-DB node based on the configuration file in the current working directory,
$ eris-db serve myChainId --work-dir=/path/to/config -- will start the Eris-DB node based on the configuration file provided and assert the chain id matches.`,
  Run: func(cmd *cobra.Command, args []string) {
    serve()
  },
}

// build the serve subcommand
func buildServeCommand() {
  addServeFlags()
}

func addServeFlags() {
  ServeCmd.PersistentFlags().StringVarP(&do.WorkDir, "work-dir", "w",
    defaultWorkDir(), "specify the working directory for the chain to run.  If omitted, and no path set in $ERIS_DB_WORKDIR, the current working directory is taken.")
}

//------------------------------------------------------------------------------
// functions

func serve() {
  // load configuration from a single location to avoid a wrong configuration
  // file is loaded.
  if err := do.ReadConfig(do.WorkDir, "server_config", "toml"); err != nil {
    log.Fatalf("Fatal error reading server_config.toml : %s \n work directory: %s \n",
      err, do.WorkDir)
    os.Exit(1)
  }
  // load chain_id for assertion
  if do.ChainId = do.Config.GetString("chain.chain_id"); do.ChainId == "" {
    log.Fatalf("Failed to read non-empty string for ChainId from config.")
    os.Exit(1)
  }

  log.Info("Eris-DB serve initializing ", do.ChainId, " from ", do.WorkDir)
}


//------------------------------------------------------------------------------
// Defaults

func defaultWorkDir() string {
  // if ERIS_DB_WORKDIR environment variable is not set, keep do.WorkDir empty
  return setDefaultString("ERIS_DB_WORKDIR", "")
}
