package templates

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/eris-ltd/eris-db/manager/eris-mint/evm"
)

const contractTemplateText = `/**
[[.Comment]]
* @dev These functions can be accessed as if this contract were deployed at the address [[.Address]]
*/
contract [[.Name]] {[[range .Functions]]
[[.SolidityIndent 1]]
[[end]]}
`
const functionTemplateText = `/**
[[.Comment]]
*/
function [[.Name]]([[.ArgList]]) constant returns ([[.Return.Type]] [[.Return.Name]]);`

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
		panic(fmt.Errorf("Couldn't parse SNative function template: %s", err))
	}
	contractTemplate, err = template.New("SolidityContractTemplate").
		Delims("[[", "]]").
		Parse(contractTemplateText)
	if err != nil {
		panic(fmt.Errorf("Couldn't parse SNative contract template: %s", err))
	}
}

type solidityContract struct {
	*vm.SNativeContractDescription
}

type solidityFunction struct {
	*vm.SNativeFunctionDescription
}

// Create a templated solidityContract from an SNative contract description
func NewSolidityContract(contract *vm.SNativeContractDescription) *solidityContract {
	return &solidityContract{contract}
}

func (contract *solidityContract) Address() string {
	return fmt.Sprintf("0x%x",
		contract.SNativeContractDescription.Address().Postfix(20))
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

// Create a templated solidityFunction from an SNative function description
func NewSolidityFunction(function *vm.SNativeFunctionDescription) *solidityFunction {
	return &solidityFunction{function}
}

func (function *solidityFunction) ArgList() string {
	argList := make([]string, len(function.Args))
	for i, arg := range function.Args {
		argList[i] = fmt.Sprintf("%s %s", arg.Type, arg.Name)
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

func (contract *solidityContract) Comment() string {
	return comment(contract.SNativeContractDescription.Comment)
}

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
