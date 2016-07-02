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
  "path"
  "time"

  tendermintConfig "github.com/tendermint/go-config"
  viper            "github.com/spf13/viper"

  config "github.com/eris-ltd/eris-db/config"
)

// NOTE [ben] Compiler check to ensure TendermintConfig successfully implements
// tendermint/go-config/config.Config
var _ tendermintConfig.Config = (*TendermintConfig)(nil)

// Tendermint has a self-rolled configuration type defined
// in tendermint/go-config but over an interface type, which is implemented
// by default in tendermint/tendermint/config/tendermint.go
// However, for Eris-DB purposes we can choose different rules for how to load
// the tendermint configuration and set the defaults.  Hence we re-implement
// go-config.Config on a viper subtree of the loaded Eris-DB configuration file.
type TendermintConfig struct {
  subTree *viper.Viper
}

func GetTendermintConfig(loadedConfig *viper.Viper) *TendermintConfig {
  // ensure we make an explicit copy
  subTree := new(viper.Viper)
  *subTree = *loadedConfig

  return &TendermintConfig {
    subTree : subTree,
  }
}

//------------------------------------------------------------------------------
// Tendermint defaults

func (tmintConfig *TendermintConfig) AssertTendermintDefaults(chainId, workDir,
  dataDir, rootDir string) {

  tmintConfig.Set("chain_id", chainId)
  tmintConfig.SetDefault("genesis_file", path.Join(rootDir, "genesis.json"))
  tmintConfig.SetDefault("proxy_app", "tcp://127.0.0.1:46658")
	tmintConfig.SetDefault("moniker", "anonymous_marmot")
	tmintConfig.SetDefault("node_laddr", "0.0.0.0:46656")
	tmintConfig.SetDefault("seeds", "")

  tmintConfig.SetDefault("fast_sync", true)
	tmintConfig.SetDefault("skip_upnp", false)
	tmintConfig.SetDefault("addrbook_file", path.Join(rootDir, "addrbook.json"))
	tmintConfig.SetDefault("priv_validator_file", path.Join(rootDir, "priv_validator.json"))
	tmintConfig.SetDefault("db_backend", "leveldb")
	tmintConfig.SetDefault("db_dir", dataDir)
	tmintConfig.SetDefault("log_level", "info")
	tmintConfig.SetDefault("rpc_laddr", "0.0.0.0:46657")
	tmintConfig.SetDefault("prof_laddr", "")
	tmintConfig.SetDefault("revision_file", path.Join(workDir,"revision"))
	tmintConfig.SetDefault("cswal", path.Join(dataDir, "cswal"))
	tmintConfig.SetDefault("cswal_light", false)

	tmintConfig.SetDefault("block_size", 10000)
	tmintConfig.SetDefault("disable_data_hash", false)
	tmintConfig.SetDefault("timeout_propose", 3000)
	tmintConfig.SetDefault("timeout_propose_delta", 500)
	tmintConfig.SetDefault("timeout_prevote", 1000)
	tmintConfig.SetDefault("timeout_prevote_delta", 500)
	tmintConfig.SetDefault("timeout_precommit", 1000)
	tmintConfig.SetDefault("timeout_precommit_delta", 500)
	tmintConfig.SetDefault("timeout_commit", 1000)
	tmintConfig.SetDefault("mempool_recheck", true)
	tmintConfig.SetDefault("mempool_recheck_empty", true)
	tmintConfig.SetDefault("mempool_broadcast", true)
}

//------------------------------------------------------------------------------
// Tendermint consistency checks

func(tmintConfig *TendermintConfig) AssertTendermintConsistency(
  consensusConfig *config.ModuleConfig, privateValidatorFilePath string) {

  tmintConfig.Set("chain_id", consensusConfig.ChainId)
  tmintConfig.Set("genesis_file", consensusConfig.GenesisFile)
  // private validator file
  tmintConfig.Set("priv_validator_file", privateValidatorFilePath)
}

// implement interface github.com/tendermint/go-config/config.Config
// so that `TMROOT` and config can be circumvented
func (tmintConfig *TendermintConfig) Get(key string) interface{} {
  return tmintConfig.subTree.Get(key)
}

func (tmintConfig *TendermintConfig) GetBool(key string) bool {
  return tmintConfig.subTree.GetBool(key)
}

func (tmintConfig *TendermintConfig) GetFloat64(key string) float64 {
  return tmintConfig.subTree.GetFloat64(key)
}

func (tmintConfig *TendermintConfig) GetInt(key string) int {
  return tmintConfig.subTree.GetInt(key)
}

func (tmintConfig *TendermintConfig) GetString(key string) string {
  return tmintConfig.subTree.GetString(key)
}

func (tmintConfig *TendermintConfig) GetStringSlice(key string) []string {
  return tmintConfig.subTree.GetStringSlice(key)
}

func (tmintConfig *TendermintConfig) GetTime(key string) time.Time {
  return tmintConfig.subTree.GetTime(key)
}

func (tmintConfig *TendermintConfig) GetMap(key string) map[string]interface{} {
  return tmintConfig.subTree.GetStringMap(key)
}

func (tmintConfig *TendermintConfig) GetMapString(key string) map[string]string {
  return tmintConfig.subTree.GetStringMapString(key)
}

func (tmintConfig *TendermintConfig) GetConfig(key string) tendermintConfig.Config {
	// TODO: [ben] log out a warning as this indicates a potentially breaking code
	// change from Tendermints side
	if !tmintConfig.subTree.IsSet(key) {
		return &TendermintConfig {
			subTree: viper.New(),
		}}
	return &TendermintConfig {
    subTree: tmintConfig.subTree.Sub(key),
  }
}

func (tmintConfig *TendermintConfig) IsSet(key string) bool {
  return tmintConfig.IsSet(key)
}

func (tmintConfig *TendermintConfig) Set(key string, value interface{}) {
  tmintConfig.subTree.Set(key, value)
}

func (tmintConfig *TendermintConfig) SetDefault(key string, value interface{}) {
  tmintConfig.subTree.SetDefault(key, value)
}
