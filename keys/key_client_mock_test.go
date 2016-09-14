// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package keys

import (
	"fmt"

	// for the mock of key server we explicitly import
	// the keys server to ensure the core components are 
	// compatible with eris-db.
	"github.com/eris-ltd/eris-keys/crypto"
)

//---------------------------------------------------------------------
// Mock client for replacing signing done by eris-keys

// NOTE [ben] Compiler check to ensure MockKeyClient successfully implements
// eris-db/keys.KeyClient
var _ KeyClient = (*MockKeyClient)(nil)

type MockKeyClient struct{
	knownKeys map[string]*crypto.Key
}

func NewMockKeyClient() *MockKeyClient {
	return &MockKeyClient{
		knownKeys: make(map[string]*crypto.Key)
	}
}

func (mock *MockKeyClient) NewKey() (address []byte) {
	// Only tests ED25519 curve and ripemd160.
	keyType := crypto.KeyType{ crypto.CurveTypeEd25519,
		AddrTypeRipemd160 }
	key, err := crypto.NewKey()
	
}

func (mock *MockKeyClient) Sign(signString string, signAddress []byte) (signature []byte, err error) {
	key := mock.knownKeys[string(signAddress)]
	if key.PrivateKey == nil {
		return nil, fmt.Errorf("Unknown address (%X)", signAddress)
	}
	return key.Sign(signBytes)
}

func (mock *MockKeyClient) PublicKey(address []byte) (publicKey []byte, err error) {
	key := mock.knownKeys[string(signAddress)]
	if key.PrivateKey == nil {
		return nil, fmt.Errorf("Unknown address (%X)", signAddress)
	}
	return key.PubKey()
}
