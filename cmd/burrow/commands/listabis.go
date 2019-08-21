package commands

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/evm/abi"
	cli "github.com/jawher/mow.cli"
)

func Listabis(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		dirs := cmd.StringsArg("DIR", nil, "Abi directory")

		cmd.Action = func() {
			for _, d := range *dirs {
				output.Printf("In %s\n", d)
				spec, err := abi.LoadPath(d)
				if err != nil {
					output.Printf("could not read %s: %v", d, err)
				}
				for id, e := range spec.EventsByID {
					fmt.Printf("event %s: %s\n", id, e.String())
				}
				for name, f := range spec.Functions {
					fmt.Printf("func %x: %s\n", f.FunctionID, f.String(name))
				}
			}
		}
	}
}
