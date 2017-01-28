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
	"os/signal"
	"path"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/eris-ltd/eris-db/core"
	"github.com/eris-ltd/eris-db/definitions"
	"github.com/eris-ltd/eris-db/logging"
	"github.com/eris-ltd/eris-db/logging/lifecycle"
	"github.com/eris-ltd/eris-db/util"
)

const (
	DefaultConfigBasename = "config"
	DefaultConfigType     = "toml"
)

var DefaultConfigFilename = fmt.Sprintf("%s.%s",
	DefaultConfigBasename,
	DefaultConfigType)

// build the serve subcommand
func buildServeCommand(do *definitions.Do) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Eris-DB serve starts an eris-db node with client API enabled by default.",
		Long: `Eris-DB serve starts an eris-db node with client API enabled by default.
The Eris-DB node is modularly configured for the consensus engine and application
manager.  The client API can be disabled.`,
		Example: fmt.Sprintf(`$ eris-db serve -- will start the Eris-DB node based on the configuration file "%s" in the current working directory
$ eris-db serve --work-dir <path-to-working-directory> -- will start the Eris-DB node based on the configuration file "%s" in the provided working directory
$ eris-db serve --chain-id <CHAIN_ID> -- will overrule the configuration entry assert_chain_id`,
			DefaultConfigFilename, DefaultConfigFilename),
		PreRun: func(cmd *cobra.Command, args []string) {
			// if WorkDir was not set by a flag or by $ERIS_DB_WORKDIR
			// NOTE [ben]: we can consider an `Explicit` flag that eliminates
			// the use of any assumptions while starting Eris-DB
			if do.WorkDir == "" {
				if currentDirectory, err := os.Getwd(); err != nil {
					panic(fmt.Sprintf("No directory provided and failed to get current "+
						"working directory: %v", err))
					os.Exit(1)
				} else {
					do.WorkDir = currentDirectory
				}
			}
			if !util.IsDir(do.WorkDir) {
				panic(fmt.Sprintf("Provided working directory %s is not a directory",
					do.WorkDir))
				os.Exit(1)
			}
		},
		Run: ServeRunner(do),
	}
	addServeFlags(do, cmd)
	return cmd
}

func addServeFlags(do *definitions.Do, serveCmd *cobra.Command) {
	serveCmd.PersistentFlags().StringVarP(&do.ChainId, "chain-id", "c",
		defaultChainId(), "specify the chain id to use for assertion against the genesis file or the existing state. If omitted, and no id is set in $CHAIN_ID, then assert_chain_id is used from the configuration file.")
	serveCmd.PersistentFlags().StringVarP(&do.WorkDir, "work-dir", "w",
		defaultWorkDir(), "specify the working directory for the chain to run.  If omitted, and no path set in $ERIS_DB_WORKDIR, the current working directory is taken.")
	serveCmd.PersistentFlags().StringVarP(&do.DataDir, "data-dir", "",
		defaultDataDir(), "specify the data directory.  If omitted and not set in $ERIS_DB_DATADIR, <working_directory>/data is taken.")
	serveCmd.PersistentFlags().BoolVarP(&do.DisableRpc, "disable-rpc", "",
		defaultDisableRpc(), "indicate for the RPC to be disabled. If omitted the RPC is enabled by default, unless (deprecated) $ERISDB_API is set to false.")
}

//------------------------------------------------------------------------------
// functions
func NewCoreFromDo(do *definitions.Do) (*core.Core, error) {
	// load the genesis file path
	do.GenesisFile = path.Join(do.WorkDir,
		do.Config.GetString("chain.genesis_file"))

	if do.Config.GetString("chain.genesis_file") == "" {
		return nil, fmt.Errorf("The config value chain.genesis_file is empty, " +
			"but should be set to the location of the genesis.json file.")
	}
	// Ensure data directory is set and accessible
	if err := do.InitialiseDataDirectory(); err != nil {
		return nil, fmt.Errorf("Failed to initialise data directory (%s): %v", do.DataDir, err)
	}

	loggerConfig, err := core.LoadLoggingConfigFromDo(do)
	if err != nil {
		return nil, fmt.Errorf("Failed to load logging config: %s", err)
	}

	// Create a root logger to pass through to dependencies
	logger := logging.WithScope(lifecycle.NewLoggerFromLoggingConfig(loggerConfig), "Serve")
	// Capture all logging from tendermint/tendermint and tendermint/go-*
	// dependencies
	lifecycle.CaptureTendermintLog15Output(logger)
	// And from stdlib go log
	lifecycle.CaptureStdlibLogOutput(logger)

	// if do.ChainId is not yet set, load chain_id for assertion from configuration file

	if do.ChainId == "" {
		if do.ChainId = do.Config.GetString("chain.assert_chain_id"); do.ChainId == "" {
			return nil, fmt.Errorf("The config chain.assert_chain_id is empty, " +
				"but should be set to the chain_id of the chain we are trying to run.")
		}
	}

	logging.Msg(logger, "Loading configuration for serve command",
		"chainId", do.ChainId,
		"workingDirectory", do.WorkDir,
		"dataDirectory", do.DataDir,
		"genesisFile", do.GenesisFile)

	consensusConfig, err := core.LoadConsensusModuleConfig(do)
	if err != nil {
		return nil, fmt.Errorf("Failed to load consensus module configuration: %s.", err)
	}

	managerConfig, err := core.LoadApplicationManagerModuleConfig(do)
	if err != nil {
		return nil, fmt.Errorf("Failed to load application manager module configuration: %s.", err)
	}

	logging.Msg(logger, "Modules configured",
		"consensusModule", consensusConfig.Version,
		"applicationManager", managerConfig.Version)

	return core.NewCore(do.ChainId, consensusConfig, managerConfig, logger)
}

// ServeRunner() returns a command runner that prepares the environment and sets
// up the core for Eris-DB to run. After the setup succeeds, it starts the core
// and waits for the core to terminate.
func ServeRunner(do *definitions.Do) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		// load configuration from a single location to avoid a wrong configuration
		// file is loaded.
		err := do.ReadConfig(do.WorkDir, DefaultConfigBasename, DefaultConfigType)
		if err != nil {
			util.Fatalf("Fatal error reading configuration from %s/%s", do.WorkDir,
				DefaultConfigFilename)
		}

		newCore, err := NewCoreFromDo(do)

		if err != nil {
			util.Fatalf("Failed to load core: %s", err)
		}

		if !do.DisableRpc {
			serverConfig, err := core.LoadServerConfig(do)
			if err != nil {
				util.Fatalf("Failed to load server configuration: %s.", err)
			}
			serverProcess, err := newCore.NewGatewayV0(serverConfig)
			if err != nil {
				util.Fatalf("Failed to load servers: %s.", err)
			}
			err = serverProcess.Start()
			if err != nil {
				util.Fatalf("Failed to start servers: %s.", err)
			}
			_, err = newCore.NewGatewayTendermint(serverConfig)
			if err != nil {
				util.Fatalf("Failed to start Tendermint gateway")
			}
			<-serverProcess.StopEventChannel()
		} else {
			signals := make(chan os.Signal, 1)
			signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
			fmt.Fprintf(os.Stderr, "Received %s signal. Marmots out.", <-signals)
		}
	}
}

//------------------------------------------------------------------------------
// Defaults

func defaultChainId() string {
	// if CHAIN_ID environment variable is not set, keep do.ChainId empty to read
	// assert_chain_id from configuration file
	return setDefaultString("CHAIN_ID", "")
}

func defaultWorkDir() string {
	// if ERIS_DB_WORKDIR environment variable is not set, keep do.WorkDir empty
	// as do.WorkDir is set by the PreRun
	return setDefaultString("ERIS_DB_WORKDIR", "")
}

func defaultDataDir() string {
	// As the default data directory depends on the default working directory,
	// wait setting a default value, and initialise the data directory from serve()
	return setDefaultString("ERIS_DB_DATADIR", "")
}

func defaultDisableRpc() bool {
	// we currently observe environment variable ERISDB_API (true = enable)
	// and default to enabling the RPC if it is not set.
	// TODO: [ben] deprecate ERISDB_API across the stack for 0.12.1, and only disable
	// the rpc through a command line flag --disable-rpc
	return !setDefaultBool("ERISDB_API", true)
}
