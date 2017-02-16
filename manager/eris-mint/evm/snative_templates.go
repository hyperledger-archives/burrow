package vm

import (
	"text/template"
	"fmt"
)

const snativeContractTemplateText=`/**
{{.SolidityComment}}
*/
contract {{.Name}} {
{{range .Functions}}
{{.Solidity}}
{{end}}
}
`
const snativeFuncTemplateText=`/**
{{.SolidityComment}}
*/
function {{.Name}}({{.SolidityArgList}}) constant returns ({{.Return.Type}} {{.Return.Name}});`

var snativeContractTemplate *template.Template
var snativeFuncTemplate *template.Template

func init() {
	var err error
	snativeFuncTemplate, err = template.New("snativeFuncTemplate").
	Parse(snativeFuncTemplateText)
	if err != nil {
		panic(fmt.Errorf("Couldn't parse SNative function template: %s", err))
	}
	snativeContractTemplate, err = template.New("snativeFuncTemplate").
			Parse(snativeContractTemplateText)
	if err != nil {
		panic(fmt.Errorf("Couldn't parse SNative contract template: %s", err))
	}
}