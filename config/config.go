package config

import (
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	logging_config "github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/rpc"
)

const DefaultBurrowConfigTOMLFileName = "burrow.toml"
const DefaultBurrowConfigEnvironmentVariable = "BURROW_CONFIG_JSON"
const DefaultGenesisDocJSONFileName = "genesis.json"

type BurrowConfig struct {
	// Set on startup
	ValidatorAddress    *crypto.Address `json:",omitempty" toml:",omitempty"`
	ValidatorPassphrase *string         `json:",omitempty" toml:",omitempty"`
	// From config file
	BurrowDir  string
	GenesisDoc *genesis.GenesisDoc                `json:",omitempty" toml:",omitempty"`
	Tendermint *tendermint.BurrowTendermintConfig `json:",omitempty" toml:",omitempty"`
	Execution  *execution.ExecutionConfig         `json:",omitempty" toml:",omitempty"`
	Keys       *keys.KeysConfig                   `json:",omitempty" toml:",omitempty"`
	RPC        *rpc.RPCConfig                     `json:",omitempty" toml:",omitempty"`
	Logging    *logging_config.LoggingConfig      `json:",omitempty" toml:",omitempty"`
}

func DefaultBurrowConfig() *BurrowConfig {
	return &BurrowConfig{
		BurrowDir:  ".burrow",
		Tendermint: tendermint.DefaultBurrowTendermintConfig(),
		Keys:       keys.DefaultKeysConfig(),
		RPC:        rpc.DefaultRPCConfig(),
		Execution:  execution.DefaultExecutionConfig(),
		Logging:    logging_config.DefaultNodeLoggingConfig(),
	}
}

func (conf *BurrowConfig) JSONString() string {
	return source.JSONString(conf)
}

func (conf *BurrowConfig) TOMLString() string {
	return source.TOMLString(conf)
}
