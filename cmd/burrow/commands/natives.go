// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/native"
	cli "github.com/jawher/mow.cli"

	"github.com/hyperledger/burrow/util/natives/templates"
)

// Dump native contracts
func Natives(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		contractsOpt := cmd.StringsOpt("c contracts", nil, "Contracts to generate")
		cmd.Action = func() {
			callables := native.MustDefaultNatives().Callables()
			// Index of next contract
			i := 1
			for _, callable := range callables {
				contract, ok := callable.(*native.Contract)
				if !ok {
					// For now we will omit loose functions (i.e. precompiles)
					// They can't be called via Solidity interface contract anyway.
					continue
				}
				if len(*contractsOpt) > 0 {
					found := false
					for _, c := range *contractsOpt {
						if c == contract.Name {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}
				solidity, err := templates.NewSolidityContract(contract).Solidity()
				if err != nil {
					fmt.Printf("Error generating solidity for contract %s: %s\n",
						contract.Name, err)
				}
				fmt.Println(solidity)
				if i < len(callables) {
					// Two new lines between contracts as per Solidity style guide
					// (the template gives us 1 trailing new line)
					fmt.Println()
				}
				i++
			}
		}
	}
}
