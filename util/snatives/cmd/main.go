package main

import (
	"fmt"

	"github.com/eris-ltd/eris-db/manager/eris-mint/evm"
	"github.com/eris-ltd/eris-db/util/snatives/templates"
)

// Dump SNative contracts
func main() {
	for _, contract := range vm.SNativeContracts() {
		solidity, err := templates.NewSolidityContract(contract).Solidity()
		if err != nil {
			fmt.Printf("Error generating solidity for contract %s: %s\n",
				contract.Name, err)
		}
		fmt.Println(solidity)
	}
}
