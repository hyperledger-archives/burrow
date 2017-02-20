package templates

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
	"github.com/eris-ltd/eris-db/manager/eris-mint/evm"
)

func TestSNativeFuncTemplate(t *testing.T) {
	contract := vm.SNativeContracts()["permissions_contract"]
	function, err := contract.FunctionByName("rm_role")
	if err != nil {
		t.Fatal("Couldn't get function")
	}
	solidityFunction := NewSolidityFunction(function)
	solidity, err := solidityFunction.Solidity()
	assert.NoError(t, err)
	fmt.Println(solidity)
}

// This test checks that we can generate the SNative contract interface and
// prints it to stdout
func TestSNativeContractTemplate(t *testing.T) {
	contract := vm.SNativeContracts()["permissions_contract"]
	solidityContract := NewSolidityContract(contract)
	solidity, err := solidityContract.Solidity()
	assert.NoError(t, err)
	fmt.Println(solidity)
}