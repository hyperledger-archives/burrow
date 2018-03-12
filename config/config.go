package config

import (
	"fmt"

	"context"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/consensus/tendermint/validator"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/crypto/inmemory"
	"github.com/hyperledger/burrow/crypto/keys"
	"github.com/hyperledger/burrow/genesis"
	logging_config "github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/burrow/rpc"
)

const DefaultBurrowConfigTOMLFileName = "burrow.toml"
const DefaultBurrowConfigJSONEnvironmentVariable = "BURROW_CONFIG_JSON"
const DefaultGenesisDocJSONFileName = "genesis.json"

type BurrowConfig struct {
	GenesisDoc *genesis.GenesisDoc                `json:",omitempty" toml:",omitempty"`
	Tendermint *tendermint.BurrowTendermintConfig `json:",omitempty" toml:",omitempty"`
	Crypto     *crypto.CryptoConfig               `json:",omitempty" toml:",omitempty"`
	RPC        *rpc.RPCConfig                     `json:",omitempty" toml:",omitempty"`
	Logging    *logging_config.LoggingConfig      `json:",omitempty" toml:",omitempty"`
}

func DefaultBurrowConfig() *BurrowConfig {
	return &BurrowConfig{
		Tendermint: tendermint.DefaultBurrowTendermintConfig(),
		Crypto:     crypto.DefaultCryptoConfig(),
		RPC:        rpc.DefaultRPCConfig(),
		Logging:    logging_config.DefaultNodeLoggingConfig(),
	}
}

func (conf *BurrowConfig) Kernel(ctx context.Context) (*core.Kernel, error) {
	if conf.GenesisDoc == nil {
		return nil, fmt.Errorf("no GenesisDoc defined in config, cannot make Kernel")
	}
	logger, err := lifecycle.NewLoggerFromLoggingConfig(conf.Logging)
	if err != nil {
		return nil, fmt.Errorf("could not generate logger from logging config: %v", err)
	}
	if conf.Crypto == nil {
		return nil, fmt.Errorf("no crypto in config, cannot make Kernel")
	}

	var val acm.Addressable
	var signer acm.Signer
	if conf.Crypto.InMemoryCrypto != nil {
		signer, err = inmemory.Signer(conf.Crypto.InMemoryCrypto)
		if err != nil {
			return nil, fmt.Errorf("could not create inmemory signer from config: %v", err)
		}
		val, err = inmemory.Addressable(conf.Crypto.InMemoryCrypto)
		if err != nil {
			return nil, fmt.Errorf("could not get validator addressable inmemory signer: %v", err)
		}
	} else if conf.Crypto.KeysServer != nil {
		if conf.Crypto.KeysServer.ValidatorAddress == nil {
			return nil, fmt.Errorf("no validator address in config, cannot make Kernel")
		}
		keyClient := keys.NewKeyClient(conf.Crypto.KeysServer.URL, logger)
		val, err = keys.Addressable(keyClient, *conf.Crypto.KeysServer.ValidatorAddress)
		if err != nil {
			return nil, fmt.Errorf("could not get validator addressable from keys client: %v", err)
		}
		signer = keys.Signer(keyClient, val.Address())
	} else {
		return nil, fmt.Errorf("no crypto provider in config, cannot make Kernel")
	}
	privValidator := validator.NewPrivValidatorMemory(val, signer)

	return core.NewKernel(ctx, privValidator, conf.GenesisDoc, conf.Tendermint.TendermintConfig(), conf.RPC, logger)
}

func (conf *BurrowConfig) JSONString() string {
	return source.JSONString(conf)
}

func (conf *BurrowConfig) TOMLString() string {
	return source.TOMLString(conf)
}
