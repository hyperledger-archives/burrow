// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package templates

import (
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/execution/native"
	"github.com/stretchr/testify/assert"
)

func TestSNativeFuncTemplate(t *testing.T) {
	contract := native.MustDefaultNatives().GetContract("Permissions")
	function := contract.FunctionByName("removeRole")
	if function == nil {
		t.Fatal("Couldn't get function")
	}
	solidityFunction := NewSolidityFunction(function)
	solidity, err := solidityFunction.Solidity()
	assert.NoError(t, err)
	fmt.Println(solidity)
}

// This test checks that we can generate the native contract interface and
// prints it to stdout
func TestSNativeContractTemplate(t *testing.T) {
	contract := native.MustDefaultNatives().GetContract("Permissions")
	solidityContract := NewSolidityContract(contract)
	solidity, err := solidityContract.Solidity()
	assert.NoError(t, err)
	fmt.Println(solidity)
}
