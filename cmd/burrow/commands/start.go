package commands

import (
	"context"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/config/source"
	logging_config "github.com/hyperledger/burrow/logging/config"
	"github.com/jawher/mow.cli"
)

func Start(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		genesisOpt := cmd.StringOpt("g genesis", "",
			"Use the specified genesis JSON file rather than a key in the main config, use - to read from STDIN")

		configOpt := cmd.StringOpt("c config", "", "Use the a specified burrow config file")

		validatorIndexOpt := cmd.Int(cli.IntOpt{
			Name:   "v validator-index",
			Desc:   "Validator index (in validators list - GenesisSpec or GenesisDoc) from which to set ValidatorAddress",
			Value:  -1,
			EnvVar: "BURROW_VALIDATOR_INDEX",
		})

		validatorAddressOpt := cmd.String(cli.StringOpt{
			Name:   "a validator-address",
			Desc:   "The address of the the signing key of this validator",
			EnvVar: "BURROW_VALIDATOR_ADDRESS",
		})

		validatorPassphraseOpt := cmd.String(cli.StringOpt{
			Name:   "p validator-passphrase",
			Desc:   "The passphrase of the signing key of this validator (currently unimplemented but planned for future version of our KeyClient interface)",
			EnvVar: "BURROW_VALIDATOR_PASSPHRASE",
		})

		validatorMonikerOpt := cmd.String(cli.StringOpt{
			Name:   "m validator-moniker",
			Desc:   "An optional human-readable moniker to identify this validator amongst Tendermint peers in logs and status queries",
			EnvVar: "BURROW_VALIDATOR_MONIKER",
		})

		cmd.Spec = "[--config=<config file>] [--validator-moniker=<human readable moniker>] " +
			"[--validator-index=<index of validator in GenesisDoc> | --validator-address=<address of validator signing key>] " +
			"[--genesis=<genesis json file>]"

		cmd.Action = func() {

			// We need to reflect on whether this obscures where values are coming from
			conf := config.DefaultBurrowConfig()
			// We treat logging a little differently in that if anything is set for logging we will not
			// set default outputs
			conf.Logging = nil
			err := source.EachOf(
				burrowConfigProvider(*configOpt),
				source.FirstOf(
					genesisDocProvider(*genesisOpt, false),
					// Try working directory
					genesisDocProvider(config.DefaultGenesisDocJSONFileName, true)),
			).Apply(conf)

			// If no logging config was provided use the default
			if conf.Logging == nil {
				conf.Logging = logging_config.DefaultNodeLoggingConfig()
			}
			if err != nil {
				output.Fatalf("could not obtain config: %v", err)
			}

			// Which validator am I?
			if *validatorAddressOpt != "" {
				address, err := acm.AddressFromHexString(*validatorAddressOpt)
				if err != nil {
					output.Fatalf("could not read address for validator in '%s'", *validatorAddressOpt)
				}
				conf.ValidatorAddress = &address
			} else if *validatorIndexOpt > -1 {
				if conf.GenesisDoc == nil {
					output.Fatalf("Unable to set ValidatorAddress from provided validator-index since no " +
						"GenesisDoc/GenesisSpec provided.")
				}
				if *validatorIndexOpt >= len(conf.GenesisDoc.Validators) {
					output.Fatalf("validator-index of %v given but only %v validators specified in GenesisDoc",
						*validatorIndexOpt, len(conf.GenesisDoc.Validators))
				}
				conf.ValidatorAddress = &conf.GenesisDoc.Validators[*validatorIndexOpt].Address
				output.Logf("Using validator index %v (address: %s)", *validatorIndexOpt, *conf.ValidatorAddress)
			}

			if *validatorPassphraseOpt != "" {
				conf.ValidatorPassphrase = validatorPassphraseOpt
			}

			if *validatorMonikerOpt != "" {
				conf.Tendermint.Moniker = *validatorMonikerOpt
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			kern, err := conf.Kernel(ctx)
			if err != nil {
				output.Fatalf("could not create Burrow kernel: %v", err)
			}

			err = kern.Boot()
			if err != nil {
				output.Fatalf("could not boot Burrow kernel: %v", err)
			}
			kern.WaitForShutdown()
		}
	}
}
