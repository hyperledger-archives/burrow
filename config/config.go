package config

import (
	"fmt"

	"context"

	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/consensus/tendermint/validator"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	logging_config "github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/lifecycle"
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
	GenesisDoc *genesis.GenesisDoc                `json:",omitempty" toml:",omitempty"`
	Tendermint *tendermint.BurrowTendermintConfig `json:",omitempty" toml:",omitempty"`
	Execution  *execution.ExecutionConfig         `json:",omitempty" toml:",omitempty"`
	Keys       *keys.KeysConfig                   `json:",omitempty" toml:",omitempty"`
	RPC        *rpc.RPCConfig                     `json:",omitempty" toml:",omitempty"`
	Logging    *logging_config.LoggingConfig      `json:",omitempty" toml:",omitempty"`
}

func DefaultBurrowConfig() *BurrowConfig {
	return &BurrowConfig{
		Tendermint: tendermint.DefaultBurrowTendermintConfig(),
		Keys:       keys.DefaultKeysConfig(),
		RPC:        rpc.DefaultRPCConfig(),
		Logging:    logging_config.DefaultNodeLoggingConfig(),
	}
}

func (conf *BurrowConfig) Kernel(ctx context.Context) (*core.Kernel, error) {
	if conf.GenesisDoc == nil {
		return nil, fmt.Errorf("no GenesisDoc defined in config, cannot make Kernel")
	}
	if conf.ValidatorAddress == nil {
		return nil, fmt.Errorf("no validator address provided, cannot make Kernel")
	}
	logger, err := lifecycle.NewLoggerFromLoggingConfig(conf.Logging)
	if err != nil {
		return nil, fmt.Errorf("could not generate logger from logging config: %v", err)
	}
	var keyClient keys.KeyClient
	var keyStore keys.KeyStore
	if conf.Keys.RemoteAddress != "" {
		keyClient, err = keys.NewRemoteKeyClient(conf.Keys.RemoteAddress, logger)
		if err != nil {
			return nil, err
		}
	} else {
		keyStore = keys.NewKeyStore(conf.Keys.KeysDirectory, conf.Keys.AllowBadFilePermissions, logger)
		keyClient = keys.NewLocalKeyClient(keyStore, logger)
	}

	val, err := keys.AddressableSigner(keyClient, *conf.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("could not get validator addressable from keys client: %v", err)
	}
	signer, err := keys.AddressableSigner(keyClient, val.Address())
	if err != nil {
		return nil, err
	}
	privValidator := validator.NewPrivValidatorMemory(val, signer)

	var exeOptions []execution.ExecutionOption
	if conf.Execution != nil {
		exeOptions, err = conf.Execution.ExecutionOptions()
		if err != nil {
			return nil, err
		}
	}

	return core.NewKernel(ctx, keyClient, privValidator, conf.GenesisDoc, conf.Tendermint.TendermintConfig(), conf.RPC, conf.Keys,
		&keyStore, exeOptions, logger)
}

func (conf *BurrowConfig) JSONString() string {
	return source.JSONString(conf)
}

func (conf *BurrowConfig) TOMLString() string {
	return source.TOMLString(conf)
}
