package commands

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/crypto"
	cli "github.com/jawher/mow.cli"
)

type configOptions struct {
	configFileOpt      *string
	genesisFileOpt     *string
	validatorIndexOpt  *int
	accountIndexOpt    *int
	initAddressOpt     *string
	initPassphraseOpt  *string
	initMonikerOpt     *string
	externalAddressOpt *string
	nodeAddressOpt     *string
}

const configFileSpec = "[--config=<config file>]"

var configFileOption = cli.StringOpt{
	Name:   "c config",
	Desc:   "Use the specified burrow config file",
	EnvVar: "BURROW_CONFIG_FILE",
}

const genesisFileSpec = "[--genesis=<genesis json file>]"

var genesisFileOption = cli.StringOpt{
	Name:   "g genesis",
	Desc:   "Use the specified genesis JSON file rather than a key in the main config, use - to read from STDIN",
	EnvVar: "BURROW_GENESIS_FILE",
}

func addConfigOptions(cmd *cli.Cmd) *configOptions {
	spec := "[--moniker=<human readable moniker>] " +
		"[--index=<index of account in GenesisDoc> " +
		"|--validator=<index of validator in GenesisDoc> " +
		"|--address=<address of signing key>] " +
		"[--passphrase=<secret passphrase to unlock key>] " +
		"[--external-address=<hostname:port>] " +
		"[--node-address=<address of node>] " +
		configFileSpec + " " + genesisFileSpec

	cmd.Spec = strings.Join([]string{cmd.Spec, spec}, " ")
	return &configOptions{
		nodeAddressOpt: cmd.String(cli.StringOpt{
			Name:   "n node-address",
			Desc:   "Use private key from keystore as node address",
			EnvVar: "BURROW_NODE_ADDRESS",
		}),

		accountIndexOpt: cmd.Int(cli.IntOpt{
			Name:   "i index",
			Desc:   "Account index (in accounts list - GenesisSpec or GenesisDoc) from which to set Address",
			Value:  -1,
			EnvVar: "BURROW_ACCOUNT_INDEX",
		}),

		validatorIndexOpt: cmd.Int(cli.IntOpt{
			Name:   "v validator",
			Desc:   "Validator index (in validators list - GenesisSpec or GenesisDoc) from which to set Address",
			Value:  -1,
			EnvVar: "BURROW_VALIDATOR_INDEX",
		}),

		initAddressOpt: cmd.String(cli.StringOpt{
			Name:   "a address",
			Desc:   "The address of the signing key of this node",
			EnvVar: "BURROW_ADDRESS",
		}),

		initPassphraseOpt: cmd.String(cli.StringOpt{
			Name:   "p passphrase",
			Desc:   "The passphrase of the signing key of this node (currently unimplemented but planned for future version of our KeyStore)",
			EnvVar: "BURROW_PASSPHRASE",
		}),

		initMonikerOpt: cmd.String(cli.StringOpt{
			Name:   "m moniker",
			Desc:   "An optional human-readable moniker to identify this node amongst Tendermint peers in logs and status queries",
			EnvVar: "BURROW_NODE_MONIKER",
		}),

		externalAddressOpt: cmd.String(cli.StringOpt{
			Name:   "x external-address",
			Desc:   "An external address or host name provided with the port that this node will broadcast over gossip in order for other nodes to connect",
			EnvVar: "BURROW_EXTERNAL_ADDRESS",
		}),

		configFileOpt: cmd.String(configFileOption),

		genesisFileOpt: cmd.String(genesisFileOption),
	}
}

func (opts *configOptions) obtainBurrowConfig() (*config.BurrowConfig, error) {
	conf, err := obtainDefaultConfig(*opts.configFileOpt, *opts.genesisFileOpt)
	if err != nil {
		return nil, err
	}
	// Which account am I?
	conf.Address, err = accountAddress(conf, *opts.initAddressOpt, *opts.accountIndexOpt, *opts.validatorIndexOpt)
	if err != nil {
		return nil, err
	}
	if *opts.initPassphraseOpt != "" {
		conf.Passphrase = opts.initPassphraseOpt
	}
	if *opts.initMonikerOpt == "" {
		chainIDHeader := ""
		if conf.GenesisDoc != nil && conf.GenesisDoc.ChainID() != "" {
			chainIDHeader = conf.GenesisDoc.ChainID() + "_"
		}
		if conf.Address != nil {
			// Set a default moniker... since we can at this stage of config completion and it is required for start
			conf.Tendermint.Moniker = fmt.Sprintf("%sNode_%s", chainIDHeader, conf.Address)
		}
	} else {
		conf.Tendermint.Moniker = *opts.initMonikerOpt
	}
	if *opts.externalAddressOpt != "" {
		conf.Tendermint.ExternalAddress = *opts.externalAddressOpt
	}
	return conf, nil
}

// address is sourced in the following order:
// 	1. explicitly from cli
// 	2. genesis accounts (by index)
// 	3. genesis validators (by index)
// 	4. config
// 	5. genesis validator (if only one)
func accountAddress(conf *config.BurrowConfig, addressIn string, accIndex, valIndex int) (*crypto.Address, error) {
	if addressIn != "" {
		address, err := crypto.AddressFromHexString(addressIn)
		if err != nil {
			return nil, fmt.Errorf("could not read address for account in '%s': %v", addressIn, err)
		}
		return &address, nil
	} else if accIndex > -1 {
		if conf.GenesisDoc == nil {
			return nil, fmt.Errorf("unable to set Address from provided index since no " +
				"GenesisDoc/GenesisSpec provided")
		}
		if accIndex >= len(conf.GenesisDoc.Accounts) {
			return nil, fmt.Errorf("index of %v given but only %v accounts specified in GenesisDoc",
				accIndex, len(conf.GenesisDoc.Accounts))
		}
		return &conf.GenesisDoc.Accounts[accIndex].Address, nil
	} else if valIndex > -1 {
		if conf.GenesisDoc == nil {
			return nil, fmt.Errorf("unable to set Address from provided validator since no " +
				"GenesisDoc/GenesisSpec provided")
		}
		if valIndex >= len(conf.GenesisDoc.Validators) {
			return nil, fmt.Errorf("validator index of %v given but only %v validators specified in GenesisDoc",
				valIndex, len(conf.GenesisDoc.Validators))
		}
		return &conf.GenesisDoc.Validators[valIndex].Address, nil
	} else if conf.Address != nil {
		return conf.Address, nil
	} else if conf.GenesisDoc != nil && len(conf.GenesisDoc.Validators) == 1 {
		return &conf.GenesisDoc.Validators[0].Address, nil
	}
	return nil, nil
}
