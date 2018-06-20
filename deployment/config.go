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

type Validator struct {
	Name        string
	Address     crypto.Address
	NodeAddress crypto.Address
}

type Key struct {
	Name    string
	Address crypto.Address
	KeyJSON json.RawMessage
}

type Config struct {
	Keys       map[crypto.Address]Key
	Validators []Validator
	Config     *genesis.GenesisDoc
}

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
