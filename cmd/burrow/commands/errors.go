package commands

import (
	"encoding/json"

	"github.com/hyperledger/burrow/execution/errors"
	cli "github.com/jawher/mow.cli"
)

func Errors(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {

		jsonOpt := cmd.BoolOpt("j json", false, "output errors as a JSON object")

		cmd.Spec = "[ --json ]"

		cmd.Action = func() {
			if *jsonOpt {
				bs, err := json.MarshalIndent(errors.Codes, "", "\t")
				if err != nil {
					output.Fatalf("Could not marshal error codes: %w", err)
				}
				output.Printf(string(bs))
			} else {
				output.Printf(errors.Codes.String())
			}
		}
	}
}
