package deployment

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"reflect"
	"text/template"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/keys"
	"github.com/pkg/errors"
	"github.com/tmthrgd/go-hex"
)

type Package struct {
	Keys         []*keys.Key
	BurrowConfig *config.BurrowConfig
}

const DefaultDumpKeysFormat = `{
  "Keys": [<< range $index, $key := . >><< if $index>>,<< end >>
    {
      "Name": "<< $key.Name >>",
      "Address": "<< $key.Address >>",
      "PublicKey": "<< base64 $key.PublicKey >>",
      "PrivateKey": "<< base64 $key.PrivateKey >>"
    }<< end >>
  ]
}`

const HelmDumpKeysFormat = `privateKeys:<< range $key := . >>
  << $key.Address >>:
    name: << $key.Name >>
    address: << $key.Address >>
    publicKey: << base64 $key.PublicKey >>
    privateKey: << base64 $key.PrivateKey >><< end >>
  `

const KubernetesKeyDumpFormat = `keysFiles:<< range $index, $key := . >>
  key-<< printf "%03d" $index >>: << base64 $key.MonaxKeysJSON >><< end >>
keysAddresses:<< range $index, $key := . >>
  key-<< printf "%03d" $index >>: << $key.Address >><< end >>
validatorAddresses:<< range $index, $key := . >>
  - << $key.Address >><< end >>
`

const LeftTemplateDelim = "<<"
const RightTemplateDelim = ">>"

var templateFuncs template.FuncMap = map[string]interface{}{
	"base64": func(rv reflect.Value) string {
		return encode(rv, base64.StdEncoding.EncodeToString)
	},
	"hex": func(rv reflect.Value) string {
		return encode(rv, hex.EncodeUpperToString)
	},
}

var DefaultDumpKeysTemplate = template.Must(template.New("MockKeyClient_DumpKeys").Funcs(templateFuncs).
	Delims(LeftTemplateDelim, RightTemplateDelim).
	Parse(DefaultDumpKeysFormat))

func (pkg *Package) Dump(templateString string) (string, error) {
	tmpl, err := template.New("DumpKeys").Delims(LeftTemplateDelim, RightTemplateDelim).Funcs(templateFuncs).
		Parse(templateString)
	if err != nil {
		return "", errors.Wrap(err, "could not dump keys to template")
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, pkg.Keys)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func encode(rv reflect.Value, encoder func([]byte) string) string {
	switch rv.Kind() {
	case reflect.Slice:
		return encoder(rv.Bytes())
	case reflect.String:
		return encoder([]byte(rv.String()))
	default:
		panic(fmt.Errorf("could not convert %#v to bytes to encode", rv))
	}
}
