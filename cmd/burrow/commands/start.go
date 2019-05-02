package commands

import (
	"github.com/hyperledger/burrow/core"
	cli "github.com/jawher/mow.cli"
)

func Start(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		genesisOpt := cmd.StringOpt("g genesis", "",
			"Use the specified genesis JSON file rather than a key in the main config, use - to read from STDIN")

		configOpt := cmd.StringOpt("c config", "", "Use the specified burrow config file")

		dialOpt := cmd.StringsOpt("d dial", nil, "Dial the specified moniker on a given host address (moniker=host)")

		cmd.Spec = "[--config=<config file>] [--genesis=<genesis json file>] [--dial=<moniker=host>]"

		configOpts := addConfigOptions(cmd)

		cmd.Action = func() {
			conf, err := obtainBurrowConfig(*configOpt, *genesisOpt)
			if err != nil {
				output.Fatalf("could not obtain config: %v", err)
			}

			err = configOpts.configure(conf)
			if err != nil {
				output.Fatalf("could not update burrow config: %v", err)
			}

			if conf.ValidatorAddress == nil {
				output.Fatalf("could not finalise validator address - please provide one in config or via --validator-address")
			}

			output.Logf("Using validator address: %s", *conf.ValidatorAddress)

			kern, err := core.LoadKernelFromConfig(conf)
			if err != nil {
				output.Fatalf("could not configure Burrow kernel: %v", err)
			}

			if err = kern.Boot(); err != nil {
				output.Fatalf("could not boot Burrow kernel: %v", err)
			}

			// dialOpt should be of the form moniker=host
			if pb := *dialOpt; len(pb) > 0 {
				kern.DialPeersFromGenesis(pb)
			}

			kern.WaitForShutdown()
		}
	}
}
