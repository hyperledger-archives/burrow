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
	"path"

	cobra "github.com/spf13/cobra"

	log "github.com/eris-ltd/eris-logger"

	core "github.com/eris-ltd/eris-db/core"
	util "github.com/eris-ltd/eris-db/util"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Eris-DB serve starts an eris-db node with client API enabled by default.",
	Long: `Eris-DB serve starts an eris-db node with client API enabled by default.
The Eris-DB node is modularly configured for the consensus engine and application
manager.  The client API can be disabled.`,
	Example: `$ eris-db serve -- will start the Eris-DB node based on the configuration file "server_config.toml" in the current working directory
$ eris-db serve --work-dir <path-to-working-directory> -- will start the Eris-DB node based on the configuration file "server_config.toml" in the provided working directory`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// if WorkDir was not set by a flag or by $ERIS_DB_WORKDIR
		// NOTE [ben]: we can consider an `Explicit` flag that eliminates
		// the use of any assumptions while starting Eris-DB
		if do.WorkDir == "" {
			if currentDirectory, err := os.Getwd(); err != nil {
				log.Fatalf("No directory provided and failed to get current working directory: %v", err)
				os.Exit(1)
			} else {

				do.WorkDir = currentDirectory
			}
		}
		if !util.IsDir(do.WorkDir) {
			log.Fatalf("Provided working directory %s is not a directory", do.WorkDir)
		}
	},
	Run: Serve,
}

// build the serve subcommand
func buildServeCommand() {
	addServeFlags()
}

func addServeFlags() {
	ServeCmd.PersistentFlags().StringVarP(&do.WorkDir, "work-dir", "w",
		defaultWorkDir(), "specify the working directory for the chain to run.  If omitted, and no path set in $ERIS_DB_WORKDIR, the current working directory is taken.")
	ServeCmd.PersistentFlags().StringVarP(&do.DataDir, "data-dir", "a",
		defaultDataDir(), "specify the data directory.  If omitted and not set in $ERIS_DB_DATADIR, <working_directory>/data is taken.")
}

//------------------------------------------------------------------------------
// functions

// serve() prepares the environment and sets up the core for Eris_DB to run.
// After the setup succeeds, serve() starts the core and halts for core to
// terminate.
func Serve(cmd *cobra.Command, args []string) {
	// load configuration from a single location to avoid a wrong configuration
	// file is loaded.
	if err := do.ReadConfig(do.WorkDir, "server_config", "toml"); err != nil {
		log.WithFields(log.Fields{
			"directory": do.WorkDir,
			"file":      "server_config.toml",
		}).Fatalf("Fatal error reading configuration")
		os.Exit(1)
	}
	// load chain_id for assertion
	if do.ChainId = do.Config.GetString("chain.assert_chain_id"); do.ChainId == "" {
		log.Fatalf("Failed to read non-empty string for ChainId from config.")
		os.Exit(1)
	}
	// load the genesis file path
	do.GenesisFile = path.Join(do.WorkDir,
		do.Config.GetString("chain.genesis_file"))
	if do.Config.GetString("chain.genesis_file") == "" {
		log.Fatalf("Failed to read non-empty string for genesis file from config.")
		os.Exit(1)
	}
	// Ensure data directory is set and accessible
	if err := do.InitialiseDataDirectory(); err != nil {
		log.Fatalf("Failed to initialise data directory (%s): %v", do.DataDir, err)
		os.Exit(1)
	}
	log.WithFields(log.Fields{
		"chainId":          do.ChainId,
		"workingDirectory": do.WorkDir,
		"dataDirectory":    do.DataDir,
		"genesisFile":      do.GenesisFile,
	}).Info("Eris-DB serve configuring")

	consensusConfig, err := core.LoadConsensusModuleConfig(do)
	if err != nil {
		log.Fatalf("Failed to load consensus module configuration: %s.", err)
		os.Exit(1)
	}

	managerConfig, err := core.LoadApplicationManagerModuleConfig(do)
	if err != nil {
		log.Fatalf("Failed to load application manager module configuration: %s.", err)
		os.Exit(1)
	}
	log.WithFields(log.Fields{
		"consensusModule":    consensusConfig.Version,
		"applicationManager": managerConfig.Version,
	}).Debug("Modules configured")

	newCore, err := core.NewCore(do.ChainId, consensusConfig, managerConfig)
	if err != nil {
		log.Fatalf("Failed to load core: %s", err)
	}

	serverConfig, err := core.LoadServerConfig(do)
	if err != nil {
		log.Fatalf("Failed to load server configuration: %s.", err)
		os.Exit(1)
	}
	serverProcess, err := newCore.NewGatewayV0(serverConfig)
	if err != nil {
		log.Fatalf("Failed to load servers: %s.", err)
		os.Exit(1)
	}
	err = serverProcess.Start()
	if err != nil {
		log.Fatalf("Failed to start servers: %s.", err)
		os.Exit(1)
	}
	_, err = newCore.NewGatewayTendermint(serverConfig)
	if err != nil {
		log.Fatalf("Failed to start Tendermint gateway")
	}
	<-serverProcess.StopEventChannel()
}

//------------------------------------------------------------------------------
// Defaults

func defaultWorkDir() string {
	// if ERIS_DB_WORKDIR environment variable is not set, keep do.WorkDir empty
	return setDefaultString("ERIS_DB_WORKDIR", "")
}

func defaultDataDir() string {
	// As the default data directory depends on the default working directory,
	// wait setting a default value, and initialise the data directory from serve()
	return setDefaultString("ERIS_DB_DATADIR", "")
}
