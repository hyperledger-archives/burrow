package main

import (
	"fmt"
	"os"

	"github.com/jawher/mow.cli"
)

func main() {
	bos := cli.App("Burrow",
		"Deep in the Burrow")
	bos.Action = func() {
		//logger, _ := lifecycle.NewStdErrLogger()
		//tendermint.LaunchGenesisValidator(logger)
	}
	bos.Run(os.Args)
}

// Print informational output to Stderr
func printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
