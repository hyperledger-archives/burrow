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

package templates

import (
	"fmt"
	"github.com/hyperledger/burrow/manager/burrow-mint/evm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSNativeFuncTemplate(t *testing.T) {
	contract := vm.SNativeContracts()["Permissions"]
	function, err := contract.FunctionByName("removeRole")
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
	contract := vm.SNativeContracts()["Permissions"]
	solidityContract := NewSolidityContract(contract)
	solidity, err := solidityContract.Solidity()
	assert.NoError(t, err)
	fmt.Println(solidity)
}
