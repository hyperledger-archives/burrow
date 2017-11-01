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
	"bytes"
	"encoding/hex"
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/tendermint/go-crypto"
)

type KeyClient interface {
	// Sign returns the signature bytes for given message signed with the key associated with signAddress
	Sign(signAddress acm.Address, message []byte) (signature crypto.Signature, err error)

	// PublicKey returns the public key associated with a given address
	PublicKey(address acm.Address) (publicKey crypto.PubKey, err error)

	// Generate requests that a key be generate within the keys instance and returns the address
	Generate(keyName string, keyType KeyType) (keyAddress acm.Address, err error)
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

// NOTE [ben] Compiler check to ensure monaxKeyClient successfully implements
// burrow/keys.KeyClient
var _ KeyClient = (*monaxKeyClient)(nil)

type monaxKeyClient struct {
	rpcString string
	logger    logging_types.InfoTraceLogger
}

type monaxSigner struct {
	keyClient KeyClient
	address   acm.Address
}

// Creates a Signer that assumes the address holds an Ed25519 key
func Signer(keyClient KeyClient, address acm.Address) acm.Signer {
	// TODO: we can do better than this and return a typed signature when we reform the keys service
	return &monaxSigner{
		keyClient: keyClient,
		address:   address,
	}
}

type keyAddressable struct {
	pubKey  crypto.PubKey
	address acm.Address
}

func (ka *keyAddressable) Address() acm.Address {
	return ka.address
}

func (ka *keyAddressable) PubKey() crypto.PubKey {
	return ka.pubKey
}

func Addressable(keyClient KeyClient, address acm.Address) (acm.Addressable, error) {
	pubKey, err := keyClient.PublicKey(address)
	if err != nil {
		return nil, err
	}
	return &keyAddressable{
		address: address,
		pubKey:  pubKey,
	}, nil
}

func (ms *monaxSigner) Sign(messsage []byte) (crypto.Signature, error) {
	signature, err := ms.keyClient.Sign(ms.address, messsage)
	if err != nil {
		return crypto.Signature{}, err
	}
	return signature, nil
}

// monaxKeyClient.New returns a new monax-keys client for provided rpc location
// Monax-keys connects over http request-responses
func NewBurrowKeyClient(rpcString string, logger logging_types.InfoTraceLogger) *monaxKeyClient {
	return &monaxKeyClient{
		rpcString: rpcString,
		logger:    logging.WithScope(logger, "BurrowKeyClient"),
	}
}

// Monax-keys client Sign requests the signature from BurrowKeysClient over rpc for the given
// bytes to be signed and the address to sign them with.
func (monaxKeys *monaxKeyClient) Sign(signAddress acm.Address, message []byte) (crypto.Signature, error) {
	args := map[string]string{
		"msg":  hex.EncodeToString(message),
		"addr": signAddress.String(),
	}
	sigS, err := RequestResponse(monaxKeys.rpcString, "sign", args, monaxKeys.logger)
	if err != nil {
		return crypto.Signature{}, err
	}
	sigBytes, err := hex.DecodeString(sigS)
	if err != nil {
		return crypto.Signature{}, err
	}
	return crypto.SignatureEd25519FromBytes(sigBytes), err
}

// Monax-keys client PublicKey requests the public key associated with an address from
// the monax-keys server.
func (monaxKeys *monaxKeyClient) PublicKey(address acm.Address) (crypto.PubKey, error) {
	args := map[string]string{
		"addr": address.String(),
	}
	pubS, err := RequestResponse(monaxKeys.rpcString, "pub", args, monaxKeys.logger)
	if err != nil {
		return crypto.PubKey{}, err
	}
	pubKey := crypto.PubKeyEd25519{}
	pubKeyBytes, err := hex.DecodeString(pubS)
	if err != nil {
		return crypto.PubKey{}, err
	}
	copy(pubKey[:], pubKeyBytes)
	if !bytes.Equal(address.Bytes(), pubKey.Address()) {
		return crypto.PubKey{}, fmt.Errorf("public key %s maps to address %X but was returned for address %s",
			pubKey, pubKey.Address(), address)
	}
	return pubKey.Wrap(), nil
}

func (monaxKeys *monaxKeyClient) Generate(keyName string, keyType KeyType) (acm.Address, error) {
	args := map[string]string{
		//"auth": auth,
		"name": keyName,
		"type": keyType.String(),
	}
	addr, err := RequestResponse(monaxKeys.rpcString, "gen", args, monaxKeys.logger)
	if err != nil {
		return acm.ZeroAddress, err
	}
	addrBytes, err := hex.DecodeString(addr)
	if err != nil {
		return acm.ZeroAddress, err
	}
	return acm.AddressFromBytes(addrBytes)
}
