package commands

import (
	"github.com/hyperledger/burrow/core"
	cli "github.com/jawher/mow.cli"
)

// Restore reads a state file and saves into a runnable dir
func Restore(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		configOpts := addConfigOptions(cmd)
		silentOpt := cmd.BoolOpt("s silent", false, "If state already exists don't throw error")
		filename := cmd.StringArg("FILE", "", "Restore from this dump")
		cmd.Spec += "[--silent] [FILE]"

		cmd.Action = func() {
			conf, err := configOpts.obtainBurrowConfig()
			if err != nil {
				output.Fatalf("could not set up config: %v", err)
			}

			if err := conf.Verify(); err != nil {
				output.Fatalf("cannot continue with config: %v", err)
			}

			output.Logf("Using validator address: %s", *conf.ValidatorAddress)

			kern, err := core.NewKernel(conf.BurrowDir)
			if err != nil {
				output.Fatalf("could not create Burrow kernel: %v", err)
			}

			if err = kern.LoadLoggerFromConfig(conf.Logging); err != nil {
				output.Fatalf("could not create Burrow kernel: %v", err)
			}

			if err = kern.LoadDump(conf.GenesisDoc, *filename, *silentOpt); err != nil {
				output.Fatalf("could not create Burrow kernel: %v", err)
			}

			kern.ShutdownAndExit()
		}
	}
}
