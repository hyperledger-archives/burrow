// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"

	cli "github.com/jawher/mow.cli"

	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/util/snatives/templates"
)

// Dump SNative contracts
func Snatives(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		contractsOpt := cmd.StringsOpt("c contracts", nil, "Contracts to generate")
		cmd.Action = func() {
			contracts := evm.SNativeContracts()
			// Index of next contract
			i := 1
			for _, contract := range contracts {
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
				if i < len(contracts) {
					// Two new lines between contracts as per Solidity style guide
					// (the template gives us 1 trailing new line)
					fmt.Println()
				}
				i++
			}
		}
	}
}
