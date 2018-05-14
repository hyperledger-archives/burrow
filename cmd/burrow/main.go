package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/burrow/cmd/burrow/commands"
	"github.com/hyperledger/burrow/project"
	"github.com/jawher/mow.cli"
)

func main() {
	burrow().Run(os.Args)
}

func burrow() *cli.Cli {
	app := cli.App("burrow", "The EVM smart contract machine with Tendermint consensus")

	versionOpt := app.BoolOpt("v version", false, "Print the Burrow version")
	app.Spec = "[--version]"

	app.Action = func() {
		if *versionOpt {
			fmt.Println(project.FullVersion())
			os.Exit(0)
		}
	}

	app.Command("start", "Start a Burrow node",
		commands.Start)

	app.Command("spec", "Build a GenesisSpec that acts as a template for a GenesisDoc and the configure command",
		commands.Spec)

	app.Command("configure",
		"Create Burrow configuration by consuming a GenesisDoc or GenesisSpec, creating keys, and emitting the config",
		commands.Configure)

	return app
}
