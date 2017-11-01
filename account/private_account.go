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
	PrivKey() crypto.PrivKey
	Signer
}

type ConcretePrivateAccount struct {
	Address Address
	PubKey  crypto.PubKey
	PrivKey crypto.PrivKey
}

type concretePrivateAccountWrapper struct {
	*ConcretePrivateAccount `json:"unwrap"`
}

var _ = wire.RegisterInterface(struct{ PrivateAccount }{}, wire.ConcreteType{concretePrivateAccountWrapper{}, 0x01})

var _ PrivateAccount = concretePrivateAccountWrapper{}

func (cpaw concretePrivateAccountWrapper) Address() Address {
	return cpaw.ConcretePrivateAccount.Address
}

func (cpaw concretePrivateAccountWrapper) PubKey() crypto.PubKey {
	return cpaw.ConcretePrivateAccount.PubKey
}

func (cpaw concretePrivateAccountWrapper) PrivKey() crypto.PrivKey {
	return cpaw.ConcretePrivateAccount.PrivKey
}

func (pa ConcretePrivateAccount) PrivateAccount() concretePrivateAccountWrapper {
	return concretePrivateAccountWrapper{&pa}
}

func (pa ConcretePrivateAccount) Sign(msg []byte) (crypto.Signature, error) {
	return pa.PrivKey.Sign(msg), nil
}

func ChainSign(pa PrivateAccount, chainID string, o Signable) crypto.Signature {
	sig, _ := pa.Sign(SignBytes(chainID, o))
	return sig
}

func (pa *ConcretePrivateAccount) Generate(index int) concretePrivateAccountWrapper {
	newPrivKey := pa.PrivKey.Unwrap().(crypto.PrivKeyEd25519).Generate(index).Wrap()
	newPubKey := newPrivKey.PubKey()
	newAddress, _ := AddressFromBytes(newPubKey.Address())
	return ConcretePrivateAccount{
		Address: newAddress,
		PubKey:  newPubKey,
		PrivKey: newPrivKey,
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
	pubKey := crypto.PubKeyEd25519(*pubKeyBytes)
	address, _ := AddressFromBytes(pubKey.Address())
	privKey := crypto.PrivKeyEd25519(*privKeyBytes)
	return ConcretePrivateAccount{
		Address: address,
		PubKey:  pubKey.Wrap(),
		PrivKey: privKey.Wrap(),
	}.PrivateAccount()
}

func PrivKeyFromSecret(secret string) crypto.PrivKey {
	return crypto.GenPrivKeyEd25519FromSecret(wire.BinarySha256(secret)).Wrap()
}

// Generates a new account with private key from SHA256 hash of a secret
func GeneratePrivateAccountFromSecret(secret string) concretePrivateAccountWrapper {
	privKey := PrivKeyFromSecret(secret)
	pubKey := privKey.PubKey()
	return ConcretePrivateAccount{
		Address: MustAddressFromBytes(pubKey.Address()),
		PubKey:  pubKey,
		PrivKey: privKey,
	}.PrivateAccount()
}

func GeneratePrivateAccountFromPrivateKeyBytes(privKeyBytes []byte) concretePrivateAccountWrapper {
	if len(privKeyBytes) != 64 {
		panic(fmt.Sprintf("Expected 64 bytes but got %v", len(privKeyBytes)))
	}
	var privKeyArray [64]byte
	copy(privKeyArray[:], privKeyBytes)
	pubKeyBytes := ed25519.MakePublicKey(&privKeyArray)
	pubKey := crypto.PubKeyEd25519(*pubKeyBytes)
	address, _ := AddressFromBytes(pubKey.Address())
	privKey := crypto.PrivKeyEd25519(privKeyArray)
	return ConcretePrivateAccount{
		Address: address,
		PubKey:  pubKey.Wrap(),
		PrivKey: privKey.Wrap(),
	}.PrivateAccount()
}
