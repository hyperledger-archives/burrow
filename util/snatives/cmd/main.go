package main

import (
	"fmt"

	"github.com/eris-ltd/eris-db/manager/eris-mint/evm"
	"github.com/eris-ltd/eris-db/util/snatives/templates"
)

// Dump SNative contracts
func main() {
	contracts := vm.SNativeContracts()
	// Index of next contract
	i := 1
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
