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

// TODO: [ben] Account and PrivateAccount need to become a pure interface
// and then move the implementation to the manager types.
// Eg, Geth has its accounts, different from BurrowMint

import (
	"fmt"

	"github.com/hyperledger/burrow/common/sanity"

	"github.com/tendermint/ed25519"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

type PrivAccount struct {
	Address []byte         `json:"address"`
	PubKey  crypto.PubKey  `json:"pub_key"`
	PrivKey crypto.PrivKey `json:"priv_key"`
}

func (pA *PrivAccount) Generate(index int) *PrivAccount {
	newPrivKey := pA.PrivKey.(crypto.PrivKeyEd25519).Generate(index)
	newPubKey := newPrivKey.PubKey()
	newAddress := newPubKey.Address()
	return &PrivAccount{
		Address: newAddress,
		PubKey:  newPubKey,
		PrivKey: newPrivKey,
	}
}

func (pA *PrivAccount) Sign(chainID string, o Signable) crypto.Signature {
	return pA.PrivKey.Sign(SignBytes(chainID, o))
}

func (pA *PrivAccount) String() string {
	return fmt.Sprintf("PrivAccount{%X}", pA.Address)
}

//----------------------------------------

// Generates a new account with private key.
func GenPrivAccount() *PrivAccount {
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], crypto.CRandBytes(32))
	pubKeyBytes := ed25519.MakePublicKey(privKeyBytes)
	pubKey := crypto.PubKeyEd25519(*pubKeyBytes)
	privKey := crypto.PrivKeyEd25519(*privKeyBytes)
	return &PrivAccount{
		Address: pubKey.Address(),
		PubKey:  pubKey,
		PrivKey: privKey,
	}
}

// Generates 32 priv key bytes from secret
func GenPrivKeyBytesFromSecret(secret string) []byte {
	return wire.BinarySha256(secret) // Not Ripemd160 because we want 32 bytes.
}

// Generates a new account with private key from SHA256 hash of a secret
func GenPrivAccountFromSecret(secret string) *PrivAccount {
	privKey32 := GenPrivKeyBytesFromSecret(secret)
	privKeyBytes := new([64]byte)
	copy(privKeyBytes[:32], privKey32)
	pubKeyBytes := ed25519.MakePublicKey(privKeyBytes)
	pubKey := crypto.PubKeyEd25519(*pubKeyBytes)
	privKey := crypto.PrivKeyEd25519(*privKeyBytes)
	return &PrivAccount{
		Address: pubKey.Address(),
		PubKey:  pubKey,
		PrivKey: privKey,
	}
}

func GenPrivAccountFromPrivKeyBytes(privKeyBytes []byte) *PrivAccount {
	if len(privKeyBytes) != 64 {
		sanity.PanicSanity(fmt.Sprintf("Expected 64 bytes but got %v", len(privKeyBytes)))
	}
	var privKeyArray [64]byte
	copy(privKeyArray[:], privKeyBytes)
	pubKeyBytes := ed25519.MakePublicKey(&privKeyArray)
	pubKey := crypto.PubKeyEd25519(*pubKeyBytes)
	privKey := crypto.PrivKeyEd25519(privKeyArray)
	return &PrivAccount{
		Address: pubKey.Address(),
		PubKey:  pubKey,
		PrivKey: privKey,
	}
}
