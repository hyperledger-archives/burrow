package main

import (
	"os"

	"github.com/tendermint/merkleeyes/cmd"
	"github.com/tendermint/tmlibs/cli"
)

func main() {
	root := cli.PrepareBaseCmd(cmd.RootCmd, "ME", os.ExpandEnv("$HOME/.merkleeyes"))
	root.Execute()
}
