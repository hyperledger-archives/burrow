package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/burrow/cmd/burrow/commands"
	"github.com/hyperledger/burrow/project"
	cli "github.com/jawher/mow.cli"
)

func main() {
	// Print informational output to Stderr
	burrow(stdOutput()).Run(os.Args)
}

func burrow(output commands.Output) *cli.Cli {
	app := cli.App("burrow", "The EVM smart contract machine with Tendermint consensus")

	versionOpt := app.BoolOpt("v version", false, "Print the Burrow version")
	app.Spec = "[--version]"

	app.Action = func() {
		if *versionOpt {
			fmt.Println(project.FullVersion())
		} else {
			app.PrintHelp()
		}
	}

	app.Command("start", "Start a Burrow node",
		commands.Start(output))

	app.Command("spec", "Build a GenesisSpec that acts as a template for a GenesisDoc and the configure command",
		commands.Spec(output))

	app.Command("configure",
		"Create Burrow configuration by consuming a GenesisDoc or GenesisSpec, creating keys, and emitting the config",
		commands.Configure(output))

	app.Command("keys", "A tool for doing a bunch of cool stuff with keys",
		commands.Keys(output))

	app.Command("dump", "Dump objects from an offline Burrow .burrow directory",
		commands.Dump(output))

	app.Command("deploy", "Deploy and test contracts",
		commands.Deploy(output))

	app.Command("snatives", "Dump Solidity interface contracts for SNatives",
		commands.Snatives(output))

	return app
}

func stdOutput() *output {
	return &output{
		PrintfFunc: func(format string, args ...interface{}) {
			fmt.Fprintf(os.Stdout, format+"\n", args...)
		},
		LogfFunc: func(format string, args ...interface{}) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		},
		FatalfFunc: func(format string, args ...interface{}) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
			os.Exit(1)
		},
	}
}

type output struct {
	PrintfFunc func(format string, args ...interface{})
	LogfFunc   func(format string, args ...interface{})
	FatalfFunc func(format string, args ...interface{})
}

func (out *output) Printf(format string, args ...interface{}) {
	out.PrintfFunc(format, args...)
}

func (out *output) Logf(format string, args ...interface{}) {
	out.LogfFunc(format, args...)
}

func (out *output) Fatalf(format string, args ...interface{}) {
	out.FatalfFunc(format, args...)
}
