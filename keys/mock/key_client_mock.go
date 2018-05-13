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

package mock

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"text/template"

	"encoding/json"

	"reflect"

	acm "github.com/hyperledger/burrow/account"
	. "github.com/hyperledger/burrow/keys"
	"github.com/pkg/errors"
	"github.com/tendermint/ed25519"
	"github.com/tendermint/go-crypto"
	"github.com/tmthrgd/go-hex"
	"github.com/wayn3h0/go-uuid"
	"golang.org/x/crypto/ripemd160"
)

//---------------------------------------------------------------------
// Mock ed25510 key for mock keys client

// Simple ed25519 key structure for mock purposes with ripemd160 address
type MockKey struct {
	Name       string
	Address    acm.Address
	PublicKey  []byte
	PrivateKey []byte
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

const KubernetesKeyDumpFormat = `keysFiles:<< range $index, $key := . >>
  key-<< printf "%03d" $index >>: << base64 $key.MonaxKeyJSON >><< end >>
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

func newMockKey(name string) (*MockKey, error) {
	key := &MockKey{
		Name:       name,
		PublicKey:  make([]byte, ed25519.PublicKeySize),
		PrivateKey: make([]byte, ed25519.PrivateKeySize),
	}
	// this is a mock key, so the entropy of the source is purely
	// for testing
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	copy(key.PrivateKey[:], privateKey[:])
	copy(key.PublicKey[:], publicKey[:])

	// prepend 0x01 for ed25519 public key
	typedPublicKeyBytes := append([]byte{0x01}, key.PublicKey...)
	hasher := ripemd160.New()
	hasher.Write(typedPublicKeyBytes)
	key.Address, err = acm.AddressFromBytes(hasher.Sum(nil))
	if err != nil {
		return nil, err
	}
	if key.Name == "" {
		key.Name = key.Address.String()
	}
	return key, nil
}

func mockKeyFromPrivateAccount(privateAccount acm.PrivateAccount) *MockKey {
	_, ok := privateAccount.PrivateKey().Unwrap().(crypto.PrivKeyEd25519)
	if !ok {
		panic(fmt.Errorf("mock key client only supports ed25519 private keys at present"))
	}
	key := &MockKey{
		Name:       privateAccount.Address().String(),
		Address:    privateAccount.Address(),
		PublicKey:  privateAccount.PublicKey().RawBytes(),
		PrivateKey: privateAccount.PrivateKey().RawBytes(),
	}
	return key
}

func (mockKey *MockKey) Sign(message []byte) (acm.Signature, error) {
	var privateKey [ed25519.PrivateKeySize]byte
	copy(privateKey[:], mockKey.PrivateKey)
	return acm.SignatureFromBytes(ed25519.Sign(&privateKey, message)[:])
}

// TODO: remove after merging keys taken from there to match serialisation
type plainKeyJSON struct {
	Id         []byte
	Type       string
	Address    string
	PrivateKey []byte
}

// Returns JSON string compatible with that stored by monax-keys
func (mockKey *MockKey) MonaxKeyJSON() string {
	id, err := uuid.NewRandom()
	if err != nil {
		return errors.Wrap(err, "could not create monax key json").Error()
	}
	jsonKey := plainKeyJSON{
		Id:         []byte(id.String()),
		Address:    mockKey.Address.String(),
		Type:       string(KeyTypeEd25519Ripemd160),
		PrivateKey: mockKey.PrivateKey,
	}
	bs, err := json.Marshal(jsonKey)
	if err != nil {
		return errors.Wrap(err, "could not create monax key json").Error()
	}
	return string(bs)
}

//---------------------------------------------------------------------
// Mock client for replacing signing done by monax-keys

// Implementation assertion
var _ KeyClient = (*MockKeyClient)(nil)

type MockKeyClient struct {
	knownKeys map[acm.Address]*MockKey
}

func NewMockKeyClient(privateAccounts ...acm.PrivateAccount) *MockKeyClient {
	client := &MockKeyClient{
		knownKeys: make(map[acm.Address]*MockKey),
	}
	for _, pa := range privateAccounts {
		client.knownKeys[pa.Address()] = mockKeyFromPrivateAccount(pa)
	}
	return client
}

func (mkc *MockKeyClient) NewKey(name string) acm.Address {
	// Only tests ED25519 curve and ripemd160.
	key, err := newMockKey(name)
	if err != nil {
		panic(fmt.Sprintf("Mocked key client failed on key generation: %s", err))
	}
	mkc.knownKeys[key.Address] = key
	return key.Address
}

func (mkc *MockKeyClient) Sign(signAddress acm.Address, message []byte) (acm.Signature, error) {
	key := mkc.knownKeys[signAddress]
	if key == nil {
		return acm.Signature{}, fmt.Errorf("Unknown address (%s)", signAddress)
	}
	return key.Sign(message)
}

func (mkc *MockKeyClient) PublicKey(address acm.Address) (acm.PublicKey, error) {
	key := mkc.knownKeys[address]
	if key == nil {
		return acm.PublicKey{}, fmt.Errorf("Unknown address (%s)", address)
	}
	pubKeyEd25519 := crypto.PubKeyEd25519{}
	copy(pubKeyEd25519[:], key.PublicKey)
	return acm.PublicKeyFromGoCryptoPubKey(pubKeyEd25519.Wrap())
}

func (mkc *MockKeyClient) Generate(keyName string, keyType KeyType) (acm.Address, error) {
	return mkc.NewKey(keyName), nil
}

func (mkc *MockKeyClient) HealthCheck() error {
	return nil
}

func (mkc *MockKeyClient) DumpKeys(templateString string) (string, error) {
	tmpl, err := template.New("DumpKeys").Delims(LeftTemplateDelim, RightTemplateDelim).Funcs(templateFuncs).
		Parse(templateString)
	if err != nil {
		return "", errors.Wrap(err, "could not dump keys to template")
	}
	buf := new(bytes.Buffer)
	keys := make([]*MockKey, 0, len(mkc.knownKeys))
	for _, k := range mkc.knownKeys {
		keys = append(keys, k)
	}
	err = tmpl.Execute(buf, keys)
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
