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
	"encoding/hex"
	"fmt"

	// for the mock of key server we explicitly import
	// the keys server to ensure the core components are
	// compatible with eris-db.
	"github.com/monax/keys/crypto"

	. "github.com/monax/eris-db/keys"
)

//---------------------------------------------------------------------
// Mock client for replacing signing done by eris-keys

// NOTE [ben] Compiler check to ensure MockKeyClient successfully implements
// eris-db/keys.KeyClient
var _ KeyClient = (*MockKeyClient)(nil)

type MockKeyClient struct {
	knownKeys map[string]*crypto.Key
}

func NewMockKeyClient() *MockKeyClient {
	return &MockKeyClient{
		knownKeys: make(map[string]*crypto.Key),
	}
}

func (mock *MockKeyClient) NewKey() (address []byte) {
	// Only tests ED25519 curve and ripemd160.
	keyType := crypto.KeyType{crypto.CurveTypeEd25519,
		crypto.AddrTypeRipemd160}
	key, err := crypto.NewKey(keyType)
	if err != nil {
		panic(fmt.Sprintf("Mocked key client failed on key generation (%s): %s", keyType.String(), err))
	}
	mock.knownKeys[fmt.Sprintf("%X", key.Address)] = key
	return key.Address
}

func (mock *MockKeyClient) Sign(signBytesString string, signAddress []byte) (signature []byte, err error) {
	key := mock.knownKeys[fmt.Sprintf("%X", signAddress)]
	if key == nil {
		return nil, fmt.Errorf("Unknown address (%X)", signAddress)
	}
	signBytes, err := hex.DecodeString(signBytesString)
	if err != nil {
		return nil, fmt.Errorf("Sign bytes string is invalid hex string: %s", err.Error())
	}
	return key.Sign(signBytes)
}

func (mock *MockKeyClient) PublicKey(address []byte) (publicKey []byte, err error) {
	key := mock.knownKeys[fmt.Sprintf("%X", address)]
	if key == nil {
		return nil, fmt.Errorf("Unknown address (%X)", address)
	}
	return key.Pubkey()
}
