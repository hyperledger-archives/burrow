package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hyperledger/burrow/cmd/burrow/commands"
	"github.com/hyperledger/burrow/project"
	cli "github.com/jawher/mow.cli"
)

func main() {
	output := defaultOutput()
	err := burrow(output).Run(os.Args)
	if err != nil {
		output.Fatalf("%v", err)
	}
}

func burrow(output commands.Output) *cli.Cli {
	app := cli.App("burrow", "The EVM smart contract machine with Tendermint consensus")
	// We'll handle any errors
	app.ErrorHandling = flag.ContinueOnError

	versionOpt := app.BoolOpt("v version", false, "Print the Burrow version")
	chDirOpt := app.StringOpt("C directory", "", "Change directory before running")
	app.Spec = "[--version] [--directory=<working directory>]"

	app.Before = func() {
		if *chDirOpt != "" {
			err := os.Chdir(*chDirOpt)
			if err != nil {
				output.Fatalf("Could not change working directory to %s: %v", *chDirOpt, err)
			}
		}
	}

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

	app.Command("explore", "Dump objects from an offline Burrow .burrow directory",
		commands.Explore(output))

	app.Command("deploy", "Deploy and test contracts",
		commands.Deploy(output))

	app.Command("natives", "Dump Solidity interface contracts for Burrow native contracts",
		commands.Natives(output))

	app.Command("vent", "Start the Vent EVM event and blocks consumer service to populated databases from smart contracts",
		commands.Vent(output))

	app.Command("dump", "Dump chain state to backup",
		commands.Dump(output))

	app.Command("tx", "Submit a transaction to a burrow node",
		commands.Tx(output))

	app.Command("restore", "Restore new chain from backup",
		commands.Restore(output))

	app.Command("accounts", "List accounts and metadata",
		commands.Accounts(output))

	app.Command("abi", "List, decode and encode using ABI",
		commands.Abi(output))

	app.Command("compile", "Compile solidity files embedding the compilation results as a fixture in a Go file",
		commands.Compile(output))

	app.Command("lite", "Start a light client to read and verify block headers",
		commands.Lite(output))

	return app
}

func defaultOutput() *output {
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
