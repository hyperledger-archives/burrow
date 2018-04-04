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
	"crypto/rand"
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	. "github.com/hyperledger/burrow/keys"
	"github.com/tendermint/ed25519"
	crypto "github.com/tendermint/go-crypto"
	"golang.org/x/crypto/ripemd160"
)

//---------------------------------------------------------------------
// Mock ed25510 key for mock keys client

// Simple ed25519 key structure for mock purposes with ripemd160 address
type MockKey struct {
	Address    acm.Address
	PrivateKey [ed25519.PrivateKeySize]byte
	PublicKey  []byte
}

func newMockKey() (*MockKey, error) {
	key := &MockKey{
		PublicKey: make([]byte, ed25519.PublicKeySize),
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
	return key, nil
}

func mockKeyFromPrivateAccount(privateAccount acm.PrivateAccount) *MockKey {
	_, ok := privateAccount.PrivateKey().Unwrap().(crypto.PrivKeyEd25519)
	if !ok {
		panic(fmt.Errorf("mock key client only supports ed25519 private keys at present"))
	}
	key := &MockKey{
		Address:   privateAccount.Address(),
		PublicKey: privateAccount.PublicKey().RawBytes(),
	}
	copy(key.PrivateKey[:], privateAccount.PrivateKey().RawBytes())
	return key
}

func (mockKey *MockKey) Sign(message []byte) (acm.Signature, error) {
	return acm.SignatureFromBytes(ed25519.Sign(&mockKey.PrivateKey, message)[:])
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

func (mock *MockKeyClient) NewKey() acm.Address {
	// Only tests ED25519 curve and ripemd160.
	key, err := newMockKey()
	if err != nil {
		panic(fmt.Sprintf("Mocked key client failed on key generation: %s", err))
	}
	mock.knownKeys[key.Address] = key
	return key.Address
}

func (mock *MockKeyClient) Sign(signAddress acm.Address, message []byte) (acm.Signature, error) {
	key := mock.knownKeys[signAddress]
	if key == nil {
		return acm.Signature{}, fmt.Errorf("Unknown address (%s)", signAddress)
	}
	return key.Sign(message)
}

func (mock *MockKeyClient) PublicKey(address acm.Address) (acm.PublicKey, error) {
	key := mock.knownKeys[address]
	if key == nil {
		return acm.PublicKey{}, fmt.Errorf("Unknown address (%s)", address)
	}
	pubKeyEd25519 := crypto.PubKeyEd25519{}
	copy(pubKeyEd25519[:], key.PublicKey)
	return acm.PublicKeyFromGoCryptoPubKey(pubKeyEd25519.Wrap())
}

func (mock *MockKeyClient) Generate(keyName string, keyType KeyType) (acm.Address, error) {
	return mock.NewKey(), nil
}

func (mock *MockKeyClient) HealthCheck() error {
	return nil
}
