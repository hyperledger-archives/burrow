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
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/keys"
	"github.com/tendermint/go-crypto"
)

//---------------------------------------------------------------------
// Mock client for replacing signing done by monax-keys

// Implementation assertion
var _ keys.KeyClient = (*KeyClient)(nil)

type KeyClient struct {
	knownKeys map[acm.Address]*Key
}

func NewKeyClient(privateAccounts ...acm.PrivateAccount) *KeyClient {
	client := &KeyClient{
		knownKeys: make(map[acm.Address]*Key),
	}
	for _, pa := range privateAccounts {
		client.knownKeys[pa.Address()] = mockKeyFromPrivateAccount(pa)
	}
	return client
}

func (mkc *KeyClient) NewKey(name string) acm.Address {
	// Only tests ED25519 curve and ripemd160.
	key, err := newKey(name)
	if err != nil {
		panic(fmt.Sprintf("Mocked key client failed on key generation: %s", err))
	}
	mkc.knownKeys[key.Address] = key
	return key.Address
}

func (mkc *KeyClient) Sign(signAddress acm.Address, message []byte) (acm.Signature, error) {
	key := mkc.knownKeys[signAddress]
	if key == nil {
		return acm.Signature{}, fmt.Errorf("unknown address (%s)", signAddress)
	}
	return key.Sign(message)
}

func (mkc *KeyClient) PublicKey(address acm.Address) (acm.PublicKey, error) {
	key := mkc.knownKeys[address]
	if key == nil {
		return acm.PublicKey{}, fmt.Errorf("unknown address (%s)", address)
	}
	pubKeyEd25519 := crypto.PubKeyEd25519{}
	copy(pubKeyEd25519[:], key.PublicKey)
	return acm.PublicKeyFromGoCryptoPubKey(pubKeyEd25519.Wrap())
}

func (mkc *KeyClient) Generate(keyName string, keyType keys.KeyType) (acm.Address, error) {
	return mkc.NewKey(keyName), nil
}

func (mkc *KeyClient) HealthCheck() error {
	return nil
}

func (mkc *KeyClient) Keys() []*Key {
	var knownKeys []*Key
	for _, key := range mkc.knownKeys {
		knownKeys = append(knownKeys, key)
	}
	return knownKeys
}
