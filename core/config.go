package core

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/consensus/tendermint/abci"
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
	if conf.RemoteAddress != "" {
		kern.keyClient, err = keys.NewRemoteKeyClient(conf.RemoteAddress, kern.Logger)
		if err != nil {
			return err
		}
	} else {
		kern.keyStore = keys.NewKeyStore(conf.KeysDirectory, conf.AllowBadFilePermissions)
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
	}
	return nil
}

// LoadTendermintFromConfig loads our consensus engine into the kernel
func (kern *Kernel) LoadTendermintFromConfig(tmConf *tendermint.BurrowTendermintConfig,
	rootDir string, privVal tmTypes.PrivValidator, validator *crypto.Address) (err error) {

	if tmConf == nil || !tmConf.Enabled {
		return nil
	}

	if privVal == nil {
		val, err := keys.AddressableSigner(kern.keyClient, *validator)
		if err != nil {
			return fmt.Errorf("could not get validator addressable from keys client: %v", err)
		}
		signer, err := keys.AddressableSigner(kern.keyClient, val.GetAddress())
		if err != nil {
			return err
		}
		privVal = tendermint.NewPrivValidatorMemory(val, signer)
	}

	authorizedPeersProvider := tmConf.DefaultAuthorizedPeersProvider()
	kern.database.Stats()

	kern.nodeInfo = fmt.Sprintf("Burrow_%s_%s_ValidatorID:%X", project.History.CurrentVersion().String(),
		kern.Blockchain.ChainID(), privVal.GetPubKey().Address())
	app := abci.NewApp(kern.nodeInfo, kern.Blockchain, kern.State, kern.exeChecker, kern.exeCommitter, kern.txCodec, authorizedPeersProvider,
		kern.Panic, kern.Logger)

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
	kern.Node, err = tendermint.NewNode(tmConf.TendermintConfig(rootDir), privVal, tmGenesisDoc, app, metricsProvider, nodeKey, tmLogger)
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

	if err = kern.LoadKeysFromConfig(conf.Keys); err != nil {
		return nil, fmt.Errorf("could not configure keys: %v", err)
	}

	if err = kern.LoadExecutionOptionsFromConfig(conf.Execution); err != nil {
		return nil, fmt.Errorf("could not add execution options: %v", err)
	}

	if err = kern.LoadState(conf.GenesisDoc); err != nil {
		return nil, fmt.Errorf("could not load state: %v", err)
	}

	if err = kern.LoadTendermintFromConfig(conf.Tendermint, conf.BurrowDir, nil, conf.ValidatorAddress); err != nil {
		return nil, fmt.Errorf("could not configure Tendermint: %v", err)
	}

	kern.AddProcesses(DefaultServices(kern, conf.RPC, conf.Keys)...)
	return kern, nil
}
