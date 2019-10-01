package commands

import (
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/crypto"
	cli "github.com/jawher/mow.cli"
)

// Start launches the burrow daemon
func Start(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		configOpts := addConfigOptions(cmd)

		cmd.Action = func() {
			conf, err := configOpts.obtainBurrowConfig()
			if err != nil {
				output.Fatalf("could not set up config: %v", err)
			}

			if err := conf.Verify(); err != nil {
				output.Fatalf("cannot continue with config: %v", err)
			}

			output.Logf("Using validator address: %s", *conf.Address)
			var nodeAddress *crypto.Address

			if *configOpts.nodeAddressOpt != "" {
				address, err := crypto.AddressFromHexString(*configOpts.nodeAddressOpt)
				if err != nil {
					output.Fatalf("nodeaddress %s not a valid address", *configOpts.nodeAddressOpt)
				}
				nodeAddress = &address
			}

			kern, err := core.LoadKernelFromConfig(conf, nodeAddress)
			if err != nil {
				output.Fatalf("could not configure Burrow kernel: %v", err)
			}

			if err = kern.Boot(); err != nil {
				output.Fatalf("could not boot Burrow kernel: %v", err)
			}

			kern.WaitForShutdown()
		}
	}
}
