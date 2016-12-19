// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT
//
// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

// config.go keeps explicit structures on the runtime configuration of
// Eris-DB and all modules.  It loads these from the Viper configuration
// loaded in `definitions.Do`
package core

import (
	"fmt"
	"os"
	"path"

	"github.com/eris-ltd/eris-db/config"
	"github.com/eris-ltd/eris-db/consensus"
	"github.com/eris-ltd/eris-db/definitions"
	"github.com/eris-ltd/eris-db/logging"
	"github.com/eris-ltd/eris-db/manager"
	"github.com/eris-ltd/eris-db/server"
	"github.com/eris-ltd/eris-db/util"
	"github.com/eris-ltd/eris-db/version"
	"github.com/spf13/viper"
)

// LoadConsensusModuleConfig wraps specifically for the consensus module
func LoadConsensusModuleConfig(do *definitions.Do) (*config.ModuleConfig, error) {
	return loadModuleConfigFromDo(do, "consensus")
}

// LoadApplicationManagerModuleConfig wraps specifically for the application
// manager
func LoadApplicationManagerModuleConfig(do *definitions.Do) (*config.ModuleConfig, error) {
	return loadModuleConfigFromDo(do, "manager")
}

func loadModuleConfigFromDo(do *definitions.Do, module string) (*config.ModuleConfig, error) {
	return LoadModuleConfig(do.Config, do.WorkDir, do.DataDir,
		do.GenesisFile, do.ChainId, module)
}

// Generic Module loader for configuration information
func LoadModuleConfig(conf *viper.Viper, rootWorkDir, rootDataDir,
	genesisFile, chainId, module string) (*config.ModuleConfig, error) {
	moduleName := conf.GetString("chain." + module + ".name")
	majorVersion := conf.GetInt("chain." + module + ".major_version")
	minorVersion := conf.GetInt("chain." + module + ".minor_version")
	minorVersionString := version.MakeMinorVersionString(moduleName, majorVersion,
		minorVersion, 0)
	if !assertValidModule(module, moduleName, minorVersionString) {
		return nil, fmt.Errorf("%s module %s (%s) is not supported by %s",
			module, moduleName, minorVersionString, version.GetVersionString())
	}
	// set up the directory structure for the module inside the data directory
	workDir := path.Join(rootDataDir, conf.GetString("chain."+module+
		".relative_root"))
	if err := util.EnsureDir(workDir, os.ModePerm); err != nil {
		return nil,
			fmt.Errorf("Failed to create module root directory %s.", workDir)
	}
	dataDir := path.Join(workDir, "data")
	if err := util.EnsureDir(dataDir, os.ModePerm); err != nil {
		return nil,
			fmt.Errorf("Failed to create module data directory %s.", dataDir)
	}
	// load configuration subtree for module
	if !conf.IsSet(moduleName) {
		return nil, fmt.Errorf("Failed to read configuration section for %s",
			moduleName)
	}
	subConfig, err := config.ViperSubConfig(conf, moduleName)
	if subConfig == nil {
		return nil, fmt.Errorf("Failed to read configuration section for %s: %s",
			moduleName, err)
	}

	return &config.ModuleConfig{
		Module:      module,
		Name:        moduleName,
		Version:     minorVersionString,
		WorkDir:     workDir,
		DataDir:     dataDir,
		RootDir:     rootWorkDir, // Eris-DB's working directory
		ChainId:     chainId,
		GenesisFile: genesisFile,
		Config:      subConfig,
	}, nil
}

// LoadServerModuleConfig wraps specifically for the servers run by core
func LoadServerConfig(do *definitions.Do) (*server.ServerConfig, error) {
	// load configuration subtree for servers
	subConfig, err := config.ViperSubConfig(do.Config, "servers")
	if err != nil {
		return nil, err
	}
	serverConfig, err := server.ReadServerConfig(subConfig)
	if err != nil {
		return nil, err
	}
	serverConfig.ChainId = do.ChainId
	return serverConfig, err
}

func LoadLoggingConfig(do *definitions.Do) (*logging.LoggingConfig, error) {
	//subConfig, err := SubConfig(conf, "logging")
	loggingConfig := &logging.LoggingConfig{}
	return loggingConfig, nil
}

//------------------------------------------------------------------------------
// Helper functions

func assertValidModule(module, name, minorVersionString string) bool {
	switch module {
	case "consensus":
		return consensus.AssertValidConsensusModule(name, minorVersionString)
	case "manager":
		return manager.AssertValidApplicationManagerModule(name, minorVersionString)
	}
	return false
}
