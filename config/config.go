package config

import (
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging/config"
)

type BurrowConfig struct {
	Genesis    genesis.GenesisDoc                `toml:"genesis"`
	Tendermint tendermint.BurrowTendermintConfig `toml:"tendermint"`
	Keys       keys.KeysConfig                   `toml:"keys"`
	Logging    config.LoggingConfig              `toml:"logging"`
}
