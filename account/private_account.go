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

package account

import (
	"fmt"

	"github.com/tendermint/ed25519"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

type Signer interface {
	Sign(msg []byte) (crypto.Signature, error)
}

type PrivateAccount interface {
	Addressable
	PrivateKey() PrivateKey
	Signer
}

type ConcretePrivateAccount struct {
	Address    Address
	PublicKey  PublicKey
	PrivateKey PrivateKey
}

type concretePrivateAccountWrapper struct {
	*ConcretePrivateAccount `json:"unwrap"`
}

var _ PrivateAccount = concretePrivateAccountWrapper{}

var _ = wire.RegisterInterface(struct{ PrivateAccount }{}, wire.ConcreteType{concretePrivateAccountWrapper{}, 0x01})

func (cpaw concretePrivateAccountWrapper) Address() Address {
	return cpaw.ConcretePrivateAccount.Address
}

func (cpaw concretePrivateAccountWrapper) PublicKey() PublicKey {
	return cpaw.ConcretePrivateAccount.PublicKey
}

func (cpaw concretePrivateAccountWrapper) PrivateKey() PrivateKey {
	return cpaw.ConcretePrivateAccount.PrivateKey
}

func (pa ConcretePrivateAccount) PrivateAccount() concretePrivateAccountWrapper {
	return concretePrivateAccountWrapper{&pa}
}

func (pa ConcretePrivateAccount) Sign(msg []byte) (crypto.Signature, error) {
	return pa.PrivateKey.Sign(msg), nil
}

func ChainSign(pa PrivateAccount, chainID string, o Signable) crypto.Signature {
	sig, _ := pa.Sign(SignBytes(chainID, o))
	return sig
}

func (pa *ConcretePrivateAccount) Generate(index int) concretePrivateAccountWrapper {
	newPrivKey := PrivateKeyFromPrivKey(pa.PrivateKey.Unwrap().(crypto.PrivKeyEd25519).Generate(index).Wrap())
	newPubKey := PublicKeyFromPubKey(newPrivKey.PubKey())
	newAddress := newPubKey.Address()
	return ConcretePrivateAccount{
		Address:    newAddress,
		PublicKey:  newPubKey,
		PrivateKey: newPrivKey,
	}.PrivateAccount()
}

func (pa *ConcretePrivateAccount) String() string {
	return fmt.Sprintf("ConcretePrivateAccount{%s}", pa.Address)
}

//----------------------------------------

// Generates a new account with private key.
func GeneratePrivateAccount() concretePrivateAccountWrapper {
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], crypto.CRandBytes(32))
	pubKeyBytes := ed25519.MakePublicKey(privKeyBytes)
	publicKey := PublicKeyFromPubKey(crypto.PubKeyEd25519(*pubKeyBytes).Wrap())
	address := publicKey.Address()
	privateKey := PrivateKeyFromPrivKey(crypto.PrivKeyEd25519(*privKeyBytes).Wrap())
	return ConcretePrivateAccount{
		Address:    address,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}.PrivateAccount()
}

func PrivateKeyFromSecret(secret string) PrivateKey {
	return PrivateKeyFromPrivKey(crypto.GenPrivKeyEd25519FromSecret(wire.BinarySha256(secret)).Wrap())
}

// Generates a new account with private key from SHA256 hash of a secret
func GeneratePrivateAccountFromSecret(secret string) concretePrivateAccountWrapper {
	privKey := PrivateKeyFromSecret(secret)
	pubKey := PublicKeyFromPubKey(privKey.PubKey())
	return ConcretePrivateAccount{
		Address:    pubKey.Address(),
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}.PrivateAccount()
}

func GeneratePrivateAccountFromPrivateKeyBytes(privKeyBytes []byte) concretePrivateAccountWrapper {
	if len(privKeyBytes) != 64 {
		panic(fmt.Sprintf("Expected 64 bytes but got %v", len(privKeyBytes)))
	}
	var privKeyArray [64]byte
	copy(privKeyArray[:], privKeyBytes)
	pubKeyBytes := ed25519.MakePublicKey(&privKeyArray)
	publicKey := PublicKeyFromPubKey(crypto.PubKeyEd25519(*pubKeyBytes).Wrap())
	address := publicKey.Address()
	privateKey := PrivateKeyFromPrivKey(crypto.PrivKeyEd25519(privKeyArray).Wrap())
	return ConcretePrivateAccount{
		Address:    address,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}.PrivateAccount()
}
