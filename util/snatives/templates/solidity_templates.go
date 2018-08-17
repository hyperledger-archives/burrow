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
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/hyperledger/burrow/execution/evm"
	"github.com/iancoleman/strcase"
)

const contractTemplateText = `pragma solidity [[.SolidityPragmaVersion]];

/**
[[.Comment]]
* @dev These functions can be accessed as if this contract were deployed at a special address ([[.Address]]).
* @dev This special address is defined as the last 20 bytes of the sha3 hash of the the contract name.
* @dev To instantiate the contract use:
* @dev [[.Name]] [[.InstanceName]] = [[.Name]](address(keccak256("[[.Name]]")));
*/
contract [[.Name]] {[[range .Functions]]
[[.SolidityIndent 1]]
[[end]]}
`
const functionTemplateText = `/**
[[.Comment]]
*/
function [[.Name]]([[.ArgList]]) public constant returns ([[.RetList]]);`

// Solidity style guide recommends 4 spaces per indentation level
// (see: http://solidity.readthedocs.io/en/develop/style-guide.html)
const indentString = "    "

var contractTemplate *template.Template
var functionTemplate *template.Template

func init() {
	var err error
	functionTemplate, err = template.New("SolidityFunctionTemplate").
		Delims("[[", "]]").
		Parse(functionTemplateText)
	if err != nil {
		panic(fmt.Errorf("couldn't parse SNative function template: %s", err))
	}
	contractTemplate, err = template.New("SolidityContractTemplate").
		Delims("[[", "]]").
		Parse(contractTemplateText)
	if err != nil {
		panic(fmt.Errorf("couldn't parse SNative contract template: %s", err))
	}
}

type solidityContract struct {
	SolidityPragmaVersion string
	*evm.SNativeContractDescription
}

type solidityFunction struct {
	*evm.SNativeFunctionDescription
}

//
// Contract
//

// Create a templated solidityContract from an SNative contract description
func NewSolidityContract(contract *evm.SNativeContractDescription) *solidityContract {
	return &solidityContract{
		SolidityPragmaVersion:      ">=0.4.24",
		SNativeContractDescription: contract,
	}
}

func (contract *solidityContract) Comment() string {
	return comment(contract.SNativeContractDescription.Comment)
}

// Get a version of the contract name to be used for an instance of the contract
func (contract *solidityContract) InstanceName() string {
	// Hopefully the contract name is UpperCamelCase. If it's not, oh well, this
	// is meant to be illustrative rather than cast iron compilable
	instanceName := strings.ToLower(contract.Name[:1]) + contract.Name[1:]
	if instanceName == contract.Name {
		return "contractInstance"
	}
	return instanceName
}

func (contract *solidityContract) Address() string {
	return fmt.Sprintf("0x%s",
		contract.SNativeContractDescription.Address())
}

// Generate Solidity code for this SNative contract
func (contract *solidityContract) Solidity() (string, error) {
	buf := new(bytes.Buffer)
	err := contractTemplate.Execute(buf, contract)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (contract *solidityContract) Functions() []*solidityFunction {
	functions := contract.SNativeContractDescription.Functions()
	solidityFunctions := make([]*solidityFunction, len(functions))
	for i, function := range functions {
		solidityFunctions[i] = NewSolidityFunction(function)
	}
	return solidityFunctions
}

//
// Function
//

// Create a templated solidityFunction from an SNative function description
func NewSolidityFunction(function *evm.SNativeFunctionDescription) *solidityFunction {
	return &solidityFunction{function}
}

func (function *solidityFunction) ArgList() string {
	argList := make([]string, len(function.Abi.Inputs))
	for i, arg := range function.Abi.Inputs {
		argList[i] = fmt.Sprintf("%s %s", arg.EVM.GetSignature(), param(arg.Name))
	}
	return strings.Join(argList, ", ")
}

func (function *solidityFunction) RetList() string {
	argList := make([]string, len(function.Abi.Outputs))
	for i, arg := range function.Abi.Outputs {
		argList[i] = fmt.Sprintf("%s %s", arg.EVM.GetSignature(), param(arg.Name))
	}
	return strings.Join(argList, ", ")
}

func (function *solidityFunction) Comment() string {
	return comment(function.SNativeFunctionDescription.Comment)
}

func (function *solidityFunction) SolidityIndent(indentLevel uint) (string, error) {
	return function.solidity(indentLevel)
}

func (function *solidityFunction) Solidity() (string, error) {
	return function.solidity(0)
}

func (function *solidityFunction) solidity(indentLevel uint) (string, error) {
	buf := new(bytes.Buffer)
	iw := NewIndentWriter(indentLevel, indentString, buf)
	err := functionTemplate.Execute(iw, function)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

//
// Utility
//

func comment(comment string) string {
	commentLines := make([]string, 0, 5)
	for _, line := range strings.Split(comment, "\n") {
		trimLine := strings.TrimLeft(line, " \t\n")
		if trimLine != "" {
			commentLines = append(commentLines, trimLine)
		}
	}
	return strings.Join(commentLines, "\n")
}

func param(name string) string {
	return "_" + strcase.ToSnake(name)
}
