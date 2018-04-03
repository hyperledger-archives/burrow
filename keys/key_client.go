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

package keys

import (
	"encoding/hex"
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/logging"
)

type KeyClient interface {
	// Sign returns the signature bytes for given message signed with the key associated with signAddress
	Sign(signAddress acm.Address, message []byte) (signature acm.Signature, err error)

	// PublicKey returns the public key associated with a given address
	PublicKey(address acm.Address) (publicKey acm.PublicKey, err error)

	// Generate requests that a key be generate within the keys instance and returns the address
	Generate(keyName string, keyType KeyType) (keyAddress acm.Address, err error)

	// Returns nil if the keys instance is healthy, error otherwise
	HealthCheck() error
}

// This mirrors "github.com/monax/keys/crypto/KeyType" but since we have no use for the struct here it seems simpler
// to replicate rather than cop an import
type KeyType string

func (kt KeyType) String() string {
	return string(kt)
}

const (
	KeyTypeEd25519Ripemd160         KeyType = "ed25519,ripemd160"
	KeyTypeEd25519Ripemd160sha256           = "ed25519,ripemd160sha256"
	KeyTypeEd25519Ripemd160sha3             = "ed25519,sha3"
	KeyTypeSecp256k1Ripemd160               = "secp256k1,ripemd160"
	KeyTypeSecp256k1Ripemd160sha256         = "secp256k1,ripemd160sha256"
	KeyTypeSecp256k1Ripemd160sha3           = "secp256k1,sha3"
	KeyTypeDefault                          = KeyTypeEd25519Ripemd160
)

// NOTE [ben] Compiler check to ensure keyClient successfully implements
// burrow/keys.KeyClient
var _ KeyClient = (*keyClient)(nil)

type keyClient struct {
	requester Requester
	logger    *logging.Logger
}

type signer struct {
	keyClient KeyClient
	address   acm.Address
}

// keyClient.New returns a new monax-keys client for provided rpc location
// Monax-keys connects over http request-responses
func NewKeyClient(rpcAddress string, logger *logging.Logger) *keyClient {
	logger = logger.WithScope("NewKeyClient")
	return &keyClient{
		requester: DefaultRequester(rpcAddress, logger),
		logger:    logger,
	}
}

// Creates a Signer that assumes the address holds an Ed25519 key
func Signer(keyClient KeyClient, address acm.Address) acm.Signer {
	// TODO: we can do better than this and return a typed signature when we reform the keys service
	return &signer{
		keyClient: keyClient,
		address:   address,
	}
}

type keyAddressable struct {
	publicKey acm.PublicKey
	address   acm.Address
}

func (ka *keyAddressable) Address() acm.Address {
	return ka.address
}

func (ka *keyAddressable) PublicKey() acm.PublicKey {
	return ka.publicKey
}

func Addressable(keyClient KeyClient, address acm.Address) (acm.Addressable, error) {
	pubKey, err := keyClient.PublicKey(address)
	if err != nil {
		return nil, err
	}
	return &keyAddressable{
		address:   address,
		publicKey: pubKey,
	}, nil
}

func (ms *signer) Sign(messsage []byte) (acm.Signature, error) {
	signature, err := ms.keyClient.Sign(ms.address, messsage)
	if err != nil {
		return acm.Signature{}, err
	}
	return signature, nil
}

// Monax-keys client Sign requests the signature from BurrowKeysClient over rpc for the given
// bytes to be signed and the address to sign them with.
func (kc *keyClient) Sign(signAddress acm.Address, message []byte) (acm.Signature, error) {
	sigS, err := kc.requester("sign", map[string]string{
		"msg":  hex.EncodeToString(message),
		"addr": signAddress.String(),
	})
	if err != nil {
		return acm.Signature{}, err
	}
	sigBytes, err := hex.DecodeString(sigS)
	if err != nil {
		return acm.Signature{}, err
	}
	return acm.SignatureFromBytes(sigBytes)
}

// Monax-keys client PublicKey requests the public key associated with an address from
// the monax-keys server.
func (kc *keyClient) PublicKey(address acm.Address) (acm.PublicKey, error) {
	pubS, err := kc.requester("pub", map[string]string{
		"addr": address.String(),
	})
	if err != nil {
		return acm.PublicKey{}, err
	}
	pubKeyBytes, err := hex.DecodeString(pubS)
	if err != nil {
		return acm.PublicKey{}, err
	}
	publicKey, err := acm.PublicKeyFromBytes(pubKeyBytes)
	if err != nil {
		return acm.PublicKey{}, err
	}
	if address != publicKey.Address() {
		return acm.PublicKey{}, fmt.Errorf("public key %s maps to address %s but was returned for address %s",
			publicKey, publicKey.Address(), address)
	}
	return publicKey, nil
}

func (kc *keyClient) Generate(keyName string, keyType KeyType) (acm.Address, error) {
	addr, err := kc.requester("gen", map[string]string{
		//"auth": auth,
		"name": keyName,
		"type": keyType.String(),
	})
	if err != nil {
		return acm.ZeroAddress, err
	}
	return acm.AddressFromHexString(addr)
}

func (kc *keyClient) HealthCheck() error {
	_, err := kc.requester("name/ls", nil)
	return err
}
