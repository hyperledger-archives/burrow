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

// version provides the current Eris-DB version and a VersionIdentifier
// for the modules to identify their version with.

package tendermint

import (
  "fmt"

  node "github.com/tendermint/tendermint/node"

  config "github.com/eris-ltd/eris-db/config"
  // files  "github.com/eris-ltd/eris-db/files"
)

type TendermintNode struct {
  tmintNode *node.Node
}

func NewTendermintNode(moduleConfig *config.ModuleConfig) (*TendermintConfig, error) {
  // re-assert proper configuration for module
  if moduleConfig.Version != GetTendermintVersion().GetMinorVersionString() {
    return nil, fmt.Errorf("Version string %s did not match %s",
      moduleConfig.Version, GetTendermintVersion().GetMinorVersionString())
  }
  // loading the module has ensured the working and data directory
  // for tendermint have been created, but the config files needs
  // to be written in tendermint's root directory.
  // NOTE: [ben] as elsewhere Sub panics if config file does not have this
  // subtree. To shield in go-routine, or PR to viper.
  tendermintConfigViper := moduleConfig.Config.Sub("configuration")
  if tendermintConfigViper == nil {
    return nil, fmt.Errorf("Failed to extract Tendermint configuration subtree.")
  }
  // wrap a copy of the viper config in a tendermint/go-config interface
  tmintConfig := GetTendermintConfig(tendermintConfigViper)
  // complete the tendermint configuration with default flags
  tmintConfig.AssertTendermintDefaults(moduleConfig.ChainId,
    moduleConfig.WorkDir, moduleConfig.DataDir, moduleConfig.RootDir)

  // newNode := node.NewNode()

  //   newNode := node.NewNode()
  //   return &TendermintNode{
  //     node: newNode,
  //   }
  return nil, fmt.Errorf("PLACEHOLDER")
}


//------------------------------------------------------------------------------
// Helper functions

// func marshalConfigToDisk(filePath string, tendermintConfig *viper.Viper) error {
//
//   tendermintConfig.Unmarshal
//   // marshal interface to toml bytes
//   bytesConfig, err := toml.Marshal(tendermintConfig)
//   if err != nil {
//     return fmt.Fatalf("Failed to marshal Tendermint configuration to bytes: %v",
//       err)
//   }
//   return files.WriteAndBackup(filePath, bytesConfig)
// }
