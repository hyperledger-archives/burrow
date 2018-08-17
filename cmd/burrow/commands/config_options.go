package commands

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/crypto"
	"github.com/jawher/mow.cli"
)

type configOptions struct {
	validatorIndexOpt      *int
	validatorAddressOpt    *string
	validatorPassphraseOpt *string
	validatorMonikerOpt    *string
	externalAddressOpt     *string
}

func addConfigOptions(cmd *cli.Cmd) *configOptions {
	spec := "[--validator-moniker=<human readable moniker>] " +
		"[--validator-index=<index of validator in GenesisDoc> | --validator-address=<address of validator signing key>] " +
		"[--validator-passphrase=<secret passphrase to unlock validator key>] " +
		"[--external-address=<hostname:port>]"

	cmd.Spec = strings.Join([]string{cmd.Spec, spec}, " ")
	return &configOptions{
		validatorIndexOpt: cmd.Int(cli.IntOpt{
			Name:   "v validator-index",
			Desc:   "Validator index (in validators list - GenesisSpec or GenesisDoc) from which to set ValidatorAddress",
			Value:  -1,
			EnvVar: "BURROW_VALIDATOR_INDEX",
		}),

		validatorAddressOpt: cmd.String(cli.StringOpt{
			Name:   "a validator-address",
			Desc:   "The address of the the signing key of this validator",
			EnvVar: "BURROW_VALIDATOR_ADDRESS",
		}),

		validatorPassphraseOpt: cmd.String(cli.StringOpt{
			Name:   "p validator-passphrase",
			Desc:   "The passphrase of the signing key of this validator (currently unimplemented but planned for future version of our KeyClient interface)",
			EnvVar: "BURROW_VALIDATOR_PASSPHRASE",
		}),

		validatorMonikerOpt: cmd.String(cli.StringOpt{
			Name:   "m validator-moniker",
			Desc:   "An optional human-readable moniker to identify this validator amongst Tendermint peers in logs and status queries",
			EnvVar: "BURROW_VALIDATOR_MONIKER",
		}),

		externalAddressOpt: cmd.String(cli.StringOpt{
			Name:   "x external-address",
			Desc:   "An external address or host name provided with the port that this node will broadcast over gossip in order for other nodes to connect",
			EnvVar: "BURROW_EXTERNAL_ADDRESS",
		}),
	}
}

func (opts *configOptions) configure(conf *config.BurrowConfig) error {
	var err error
	// Which validator am I?
	conf.ValidatorAddress, err = validatorAddress(conf, *opts.validatorAddressOpt, *opts.validatorIndexOpt)
	if err != nil {
		return err
	}
	if *opts.validatorPassphraseOpt != "" {
		conf.ValidatorPassphrase = opts.validatorPassphraseOpt
	}
	if *opts.validatorMonikerOpt != "" {
		conf.Tendermint.Moniker = *opts.validatorMonikerOpt
	}
	if *opts.externalAddressOpt != "" {
		conf.Tendermint.ExternalAddress = *opts.externalAddressOpt
	}
	return nil
}

func validatorAddress(conf *config.BurrowConfig, addressString string, index int) (*crypto.Address, error) {
	if addressString != "" {
		address, err := crypto.AddressFromHexString(addressString)
		if err != nil {
			return nil, fmt.Errorf("could not read address for validator in '%s': %v", addressString, err)
		}
		return &address, nil
	} else if index > -1 {
		if conf.GenesisDoc == nil {
			return nil, fmt.Errorf("unable to set ValidatorAddress from provided validator-index since no " +
				"GenesisDoc/GenesisSpec provided")
		}
		if index >= len(conf.GenesisDoc.Validators) {
			return nil, fmt.Errorf("validator-index of %v given but only %v validators specified in GenesisDoc",
				index, len(conf.GenesisDoc.Validators))
		}
		return &conf.GenesisDoc.Validators[index].Address, nil
	}
	if conf.ValidatorAddress != nil {
		return conf.ValidatorAddress, nil
	}
	return nil, nil
}
