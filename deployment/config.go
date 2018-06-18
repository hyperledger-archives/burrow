package deployment

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"text/template"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/genesis"
	"github.com/pkg/errors"
	"github.com/tmthrgd/go-hex"
)

type Config struct {
	Config genesis.GenesisDoc
}

type Key struct {
	Name    string
	Address crypto.Address
	KeyJSON json.RawMessage
}

type KeysSecret struct {
	Keys      []Key
	NodeKeys  []Key
	ChainName string
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

const KubernetesKeyDumpFormat = `apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: << .ChainName >>-keys
data:
<<- range .Keys >>
  << .Address >>.json: << base64 .KeyJSON >>
<<- end >>
<<- range .NodeKeys >>
  << .Name >>: << base64 .KeyJSON >>
<<- end >>
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
	Parse(KubernetesKeyDumpFormat))

func (pkg *KeysSecret) Dump(templateString string) (string, error) {
	tmpl, err := template.New("DumpKeys").Delims(LeftTemplateDelim, RightTemplateDelim).Funcs(templateFuncs).
		Parse(templateString)
	if err != nil {
		return "", errors.Wrap(err, "could not dump keys to template")
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, pkg)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (pkg *Config) Dump(templateString string) (string, error) {
	tmpl, err := template.New("ConfigTemplate").Delims(LeftTemplateDelim, RightTemplateDelim).Funcs(templateFuncs).
		Parse(templateString)
	if err != nil {
		return "", errors.Wrap(err, "could not dump config to template")
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, pkg)
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
