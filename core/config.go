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

  consensus   "github.com/eris-ltd/eris-db/consensus"
  config      "github.com/eris-ltd/eris-db/config"
  definitions "github.com/eris-ltd/eris-db/definitions"
  manager     "github.com/eris-ltd/eris-db/manager"
  util        "github.com/eris-ltd/eris-db/util"
  version     "github.com/eris-ltd/eris-db/version"
)

// LoadConsensusModuleConfig wraps specifically for the consensus module
func LoadConsensusModuleConfig(do *definitions.Do) (*config.ModuleConfig, error) {
  return loadModuleConfig(do, "consensus")
}

// LoadApplicationManagerModuleConfig wraps specifically for the application
// manager
func LoadApplicationManagerModuleConfig(do *definitions.Do) (*config.ModuleConfig, error) {
  return loadModuleConfig(do, "manager")
}

// Generic Module loader for configuration information
func loadModuleConfig(do *definitions.Do, module string) (*config.ModuleConfig, error) {
  moduleName := do.Config.GetString("chain." + module + ".name")
  majorVersion := do.Config.GetInt("chain." + module + ".major_version")
  minorVersion := do.Config.GetInt("chain." + module + ".minor_version")
  minorVersionString := version.MakeMinorVersionString(moduleName, majorVersion,
    minorVersion, 0)
  if !assertValidModule(module, moduleName, minorVersionString) {
    return nil, fmt.Errorf("%s module %s (%s) is not supported by %s",
      module, moduleName, minorVersionString, version.GetVersionString())
  }
  // set up the directory structure for the module inside the data directory
  workDir := path.Join(do.DataDir, do.Config.GetString("chain." + module +
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
  // TODO: [ben] Viper internally panics if `moduleName` contains an unallowed
  // character (eg, a dash).  Either this needs to be wrapped in a go-routine
  // and recovered from or a PR to viper is needed to address this bug.
  subConfig := do.Config.Sub(moduleName)
  if subConfig == nil {
    return nil,
      fmt.Errorf("Failed to read configuration section for %s.", moduleName)
  }

  return &config.ModuleConfig {
    Module  :     module,
    Name    :     moduleName,
    Version :     minorVersionString,
    WorkDir :     workDir,
    DataDir :     dataDir,
    RootDir :     do.WorkDir, // Eris-DB's working directory
    ChainId :     do.ChainId,
    GenesisFile : do.GenesisFile,
    Config :      subConfig,
  }, nil
}

//------------------------------------------------------------------------------
// Helper functions

func assertValidModule(module, name, minorVersionString string) bool {
  switch module {
  case "consensus" :
    return consensus.AssertValidConsensusModule(name, minorVersionString)
  case "manager" :
    return manager.AssertValidApplicationManagerModule(name, minorVersionString)
  }
  return false
}
