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

type PrivateAccount interface {
	Addressable
	PrivKey() crypto.PrivKey
	Sign(chainID string, o Signable) crypto.Signature
}

type ConcretePrivateAccount struct {
	Address Address        `json:"address"`
	PubKey  crypto.PubKey  `json:"pub_key"`
	PrivKey crypto.PrivKey `json:"priv_key"`
}

type concretePrivateAccountWrapper struct {
	*ConcretePrivateAccount
}

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

func (cpaw concretePrivateAccountWrapper) Unwrap() *ConcretePrivateAccount {
	return cpaw.ConcretePrivateAccount
}

func (pa *ConcretePrivateAccount) Wrap() concretePrivateAccountWrapper {
	return concretePrivateAccountWrapper{pa}
}

func (pa ConcretePrivateAccount) Sign(chainID string, o Signable) crypto.Signature {
	return pa.PrivKey.Sign(SignBytes(chainID, o))
}


func (pa *ConcretePrivateAccount) Generate(index int) concretePrivateAccountWrapper {
	newPrivKey := pa.PrivKey.Unwrap().(crypto.PrivKeyEd25519).Generate(index).Wrap()
	newPubKey := newPrivKey.PubKey()
	newAddress, _ := AddressFromBytes(newPubKey.Address())
	cpa := &ConcretePrivateAccount{
		Address: newAddress,
		PubKey:  newPubKey,
		PrivKey: newPrivKey,
	}
	return cpa.Wrap()
}

func (pa *ConcretePrivateAccount) String() string {
	return fmt.Sprintf("ConcretePrivateAccount{%s}", pa.Address)
}

//----------------------------------------

// Generates a new account with private key.
func GenPrivAccount() concretePrivateAccountWrapper {
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], crypto.CRandBytes(32))
	pubKeyBytes := ed25519.MakePublicKey(privKeyBytes)
	pubKey := crypto.PubKeyEd25519(*pubKeyBytes)
	address, _ := AddressFromBytes(pubKey.Address())
	privKey := crypto.PrivKeyEd25519(*privKeyBytes)
	cpa := &ConcretePrivateAccount{
		Address: address,
		PubKey:  pubKey.Wrap(),
		PrivKey: privKey.Wrap(),
	}
	return cpa.Wrap()
}

// Generates 32 priv key bytes from secret
func GenPrivKeyBytesFromSecret(secret string) []byte {
	return wire.BinarySha256(secret) // Not Ripemd160 because we want 32 bytes.
}

// Generates a new account with private key from SHA256 hash of a secret
func GenPrivAccountFromSecret(secret string) concretePrivateAccountWrapper {
	privKey32 := GenPrivKeyBytesFromSecret(secret)
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], privKey32)
	pubKeyBytes := ed25519.MakePublicKey(privKeyBytes)
	pubKey := crypto.PubKeyEd25519(*pubKeyBytes)
	address, _ := AddressFromBytes(pubKey.Address())
	privKey := crypto.PrivKeyEd25519(*privKeyBytes)
	cpa := &ConcretePrivateAccount{
		Address: address,
		PubKey:  pubKey.Wrap(),
		PrivKey: privKey.Wrap(),
	}
	return cpa.Wrap()
}

func GenPrivAccountFromPrivKeyBytes(privKeyBytes []byte) concretePrivateAccountWrapper {
	if len(privKeyBytes) != 64 {
		panic(fmt.Sprintf("Expected 64 bytes but got %v", len(privKeyBytes)))
	}
	var privKeyArray [64]byte
	copy(privKeyArray[:], privKeyBytes)
	pubKeyBytes := ed25519.MakePublicKey(&privKeyArray)
	pubKey := crypto.PubKeyEd25519(*pubKeyBytes)
	address, _ := AddressFromBytes(pubKey.Address())
	privKey := crypto.PrivKeyEd25519(privKeyArray)
	cpa := &ConcretePrivateAccount{
		Address: address,
		PubKey:  pubKey.Wrap(),
		PrivKey: privKey.Wrap(),
	}
	return cpa.Wrap()
}
