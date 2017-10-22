package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/jawher/mow.cli"
)

func main() {
	burrow := cli.App("Burrow",
		"Deep in the Burrow")
	burrow.Action = func() {
		logger, _ := lifecycle.NewStdErrLogger()
		core.NewGenesisKernel()
	}
	burrow.Run(os.Args)
}

// Print informational output to Stderr
func printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
