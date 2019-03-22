package commands

import (
	"context"

	"github.com/hyperledger/burrow/core"
	cli "github.com/jawher/mow.cli"
)

func Restore(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		genesisOpt := cmd.StringOpt("g genesis", "",
			"Use the specified genesis JSON file rather than a key in the main config, use - to read from STDIN")

		configOpt := cmd.StringOpt("c config", "", "Use the specified burrow config file")

		filename := cmd.StringArg("FILE", "", "Restore from this dump")

		cmd.Spec = "[--config=<config file>] [--genesis=<genesis json file>] [FILE]"

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

			kern, err := core.NewKernel(conf.BurrowDir)
			if err != nil {
				output.Fatalf("could not create Burrow kernel: %v", err)
			}

			if err = kern.LoadLoggerFromConfig(conf.Logging); err != nil {
				output.Fatalf("could not create Burrow kernel: %v", err)
			}

			if err = kern.LoadDump(conf.GenesisDoc, *filename); err != nil {
				output.Fatalf("could not create Burrow kernel: %v", err)
			}

			kern.Shutdown(context.Background())
		}
	}
}
