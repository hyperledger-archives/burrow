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

package main

import (
	"fmt"

	"github.com/monax/burrow/manager/eris-mint/evm"
	"github.com/monax/burrow/util/snatives/templates"
)

// Dump SNative contracts
func main() {
	contracts := vm.SNativeContracts()
	// Index of next contract
	i := 1
	fmt.Print("pragma solidity >=0.0.0;\n\n")
	for _, contract := range contracts {
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
