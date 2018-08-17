package commands

import (
	"context"

	"github.com/jawher/mow.cli"
)

func Start(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		genesisOpt := cmd.StringOpt("g genesis", "",
			"Use the specified genesis JSON file rather than a key in the main config, use - to read from STDIN")

		configOpt := cmd.StringOpt("c config", "", "Use the a specified burrow config file")

		cmd.Spec = "[--config=<config file>] [--genesis=<genesis json file>]"

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
