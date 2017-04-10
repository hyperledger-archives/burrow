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

package tendermint

import (
	"path"
	"time"

	"github.com/spf13/viper"
	tendermintConfig "github.com/tendermint/go-config"

	"github.com/monax/burrow/config"
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

	return &TendermintConfig{
		subTree: subTree,
	}
}

//------------------------------------------------------------------------------
// Tendermint defaults

//
// Contract
//

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
	tmintConfig.SetDefault("addrbook_strict", true) // disable to allow connections locally
	tmintConfig.SetDefault("pex_reactor", false)    // enable for peer exchange
	tmintConfig.SetDefault("priv_validator_file", path.Join(rootDir, "priv_validator.json"))
	tmintConfig.SetDefault("db_backend", "leveldb")
	tmintConfig.SetDefault("db_dir", dataDir)
	tmintConfig.SetDefault("log_level", "info")
	tmintConfig.SetDefault("rpc_laddr", "")
	tmintConfig.SetDefault("prof_laddr", "")
	tmintConfig.SetDefault("revision_file", path.Join(workDir, "revision"))
	tmintConfig.SetDefault("cs_wal_dir", path.Join(dataDir, "cs.wal"))
	tmintConfig.SetDefault("cs_wal_light", false)
	tmintConfig.SetDefault("filter_peers", false)

	tmintConfig.SetDefault("block_size", 10000)      // max number of txs
	tmintConfig.SetDefault("block_part_size", 65536) // part size 64K
	tmintConfig.SetDefault("disable_data_hash", false)
	tmintConfig.SetDefault("timeout_propose", 3000)
	tmintConfig.SetDefault("timeout_propose_delta", 500)
	tmintConfig.SetDefault("timeout_prevote", 1000)
	tmintConfig.SetDefault("timeout_prevote_delta", 500)
	tmintConfig.SetDefault("timeout_precommit", 1000)
	tmintConfig.SetDefault("timeout_precommit_delta", 500)
	tmintConfig.SetDefault("timeout_commit", 1000)
	// make progress asap (no `timeout_commit`) on full precommit votes
	tmintConfig.SetDefault("skip_timeout_commit", false)
	tmintConfig.SetDefault("mempool_recheck", true)
	tmintConfig.SetDefault("mempool_recheck_empty", true)
	tmintConfig.SetDefault("mempool_broadcast", true)
	tmintConfig.SetDefault("mempool_wal_dir", path.Join(dataDir, "mempool.wal"))
}

//------------------------------------------------------------------------------
// Tendermint consistency checks

func (tmintConfig *TendermintConfig) AssertTendermintConsistency(
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
	subTree, _ := config.ViperSubConfig(tmintConfig.subTree, key)
	if subTree == nil {
		return &TendermintConfig{
			subTree: viper.New(),
		}
	}
	return &TendermintConfig{
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
