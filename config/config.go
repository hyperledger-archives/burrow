package config

import (
	"fmt"

	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/rpc"
	tmConfig "github.com/tendermint/tendermint/config"
)

const DefaultBurrowConfigTOMLFileName = "burrow.toml"
const DefaultBurrowConfigEnvironmentVariable = "BURROW_CONFIG_JSON"
const DefaultGenesisDocJSONFileName = "genesis.json"

type BurrowConfig struct {
	// Set on startup
	Address    *crypto.Address `json:",omitempty" toml:",omitempty"`
	Passphrase *string         `json:",omitempty" toml:",omitempty"`
	// From config file
	BurrowDir  string
	GenesisDoc *genesis.GenesisDoc                `json:",omitempty" toml:",omitempty"`
	Tendermint *tendermint.BurrowTendermintConfig `json:",omitempty" toml:",omitempty"`
	Execution  *execution.ExecutionConfig         `json:",omitempty" toml:",omitempty"`
	Keys       *keys.KeysConfig                   `json:",omitempty" toml:",omitempty"`
	RPC        *rpc.RPCConfig                     `json:",omitempty" toml:",omitempty"`
	Logging    *logconfig.LoggingConfig           `json:",omitempty" toml:",omitempty"`
}

func DefaultBurrowConfig() *BurrowConfig {
	return &BurrowConfig{
		BurrowDir:  ".burrow",
		Tendermint: tendermint.DefaultBurrowTendermintConfig(),
		Keys:       keys.DefaultKeysConfig(),
		RPC:        rpc.DefaultRPCConfig(),
		Execution:  execution.DefaultExecutionConfig(),
		Logging:    logconfig.DefaultNodeLoggingConfig(),
	}
}

func (conf *BurrowConfig) Verify() error {
	if conf.Address == nil {
		return fmt.Errorf("could not finalise address - please provide one in config or via --account-address")
	}
	return nil
}

func (conf *BurrowConfig) TendermintConfig() *tmConfig.Config {
	return conf.Tendermint.Config(conf.BurrowDir, conf.Execution.TimeoutFactor)
}

func (conf *BurrowConfig) JSONString() string {
	return source.JSONString(conf)
}

func (conf *BurrowConfig) TOMLString() string {
	return source.TOMLString(conf)
}
