// Copyright 2019 Monax Industries Limited
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

package core

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/consensus/abci"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/project"
	tmConfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	tmTypes "github.com/tendermint/tendermint/types"
)

// LoadKeysFromConfig sets the keyClient & keyStore based on the given config
func (kern *Kernel) LoadKeysFromConfig(conf *keys.KeysConfig) (err error) {
	kern.keyStore = keys.NewKeyStore(conf.KeysDirectory, conf.AllowBadFilePermissions)
	if conf.RemoteAddress != "" {
		kern.keyClient, err = keys.NewRemoteKeyClient(conf.RemoteAddress, kern.Logger)
		if err != nil {
			return err
		}
	} else {
		kern.keyClient = keys.NewLocalKeyClient(kern.keyStore, kern.Logger)
	}
	return nil
}

// LoadLoggerFromConfig adds a logging configuration to the kernel
func (kern *Kernel) LoadLoggerFromConfig(conf *logconfig.LoggingConfig) error {
	logger, err := lifecycle.NewLoggerFromLoggingConfig(conf)
	kern.SetLogger(logger)
	return err
}

// LoadExecutionOptionsFromConfig builds the execution options for the kernel
func (kern *Kernel) LoadExecutionOptionsFromConfig(conf *execution.ExecutionConfig) error {
	if conf != nil {
		exeOptions, err := conf.ExecutionOptions()
		if err != nil {
			return err
		}
		kern.exeOptions = exeOptions
		kern.timeoutFactor = conf.TimeoutFactor
	}
	return nil
}

// LoadTendermintFromConfig loads our consensus engine into the kernel
func (kern *Kernel) LoadTendermintFromConfig(conf *config.BurrowConfig, privVal tmTypes.PrivValidator) (err error) {
	if conf.Tendermint == nil || !conf.Tendermint.Enabled {
		return nil
	}

	authorizedPeersProvider := conf.Tendermint.DefaultAuthorizedPeersProvider()
	kern.database.Stats()

	kern.info = fmt.Sprintf("Burrow_%s_%s_ValidatorID:%X", project.History.CurrentVersion().String(),
		kern.Blockchain.ChainID(), privVal.GetPubKey().Address())

	app := abci.NewApp(kern.info, kern.Blockchain, kern.State, kern.checker, kern.committer, kern.txCodec,
		authorizedPeersProvider, kern.Panic, kern.Logger)

	// We could use this to provide/register our own metrics (though this will register them with us). Unfortunately
	// Tendermint currently ignores the metrics passed unless its own server is turned on.
	metricsProvider := node.DefaultMetricsProvider(&tmConfig.InstrumentationConfig{
		Prometheus:           false,
		PrometheusListenAddr: "",
	})

	genesisDoc := kern.Blockchain.GenesisDoc()

	// find node key
	var nodeKey *crypto.PrivateKey
	for _, v := range genesisDoc.Validators {
		thisAddress, err := crypto.AddressFromHexString(privVal.GetPubKey().Address().String())
		if err != nil {
			break
		}
		if v.Address == thisAddress && v.NodeAddress != nil {
			k, err := kern.keyStore.GetKey("", v.NodeAddress.Bytes())
			if err == nil {
				nodeKey = &k.PrivateKey
			}
			break
		}
	}

	tmGenesisDoc := tendermint.DeriveGenesisDoc(&genesisDoc, kern.Blockchain.AppHashAfterLastBlock())
	heightValuer := log.Valuer(func() interface{} { return kern.Blockchain.LastBlockHeight() })
	tmLogger := kern.Logger.With(structure.CallerKey, log.Caller(LoggingCallerDepth+1)).With("height", heightValuer)
	kern.Node, err = tendermint.NewNode(conf.TendermintConfig(), privVal, tmGenesisDoc, app, metricsProvider, nodeKey, tmLogger)
	return err
}

// LoadKernelFromConfig builds and returns a Kernel based solely on the supplied configuration
func LoadKernelFromConfig(conf *config.BurrowConfig) (*Kernel, error) {
	kern, err := NewKernel(conf.BurrowDir)
	if err != nil {
		return nil, fmt.Errorf("could not create initial kernel: %v", err)
	}

	if err = kern.LoadLoggerFromConfig(conf.Logging); err != nil {
		return nil, fmt.Errorf("could not configure logger: %v", err)
	}

	err = kern.LoadKeysFromConfig(conf.Keys)
	if err != nil {
		return nil, fmt.Errorf("could not configure keys: %v", err)
	}

	err = kern.LoadExecutionOptionsFromConfig(conf.Execution)
	if err != nil {
		return nil, fmt.Errorf("could not add execution options: %v", err)
	}

	err = kern.LoadState(conf.GenesisDoc)
	if err != nil {
		return nil, fmt.Errorf("could not load state: %v", err)
	}

	if conf.Address == nil {
		return nil, fmt.Errorf("Address must be set")
	}

	privVal, err := kern.PrivValidator(*conf.Address)
	if err != nil {
		return nil, fmt.Errorf("could not form PrivValidator from Address: %v", err)
	}

	err = kern.LoadTendermintFromConfig(conf, privVal)
	if err != nil {
		return nil, fmt.Errorf("could not configure Tendermint: %v", err)
	}

	kern.AddProcesses(DefaultProcessLaunchers(kern, conf.RPC, conf.Keys)...)
	return kern, nil
}
