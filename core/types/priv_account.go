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

// TODO: [ben] Account and PrivateAccount need to become a pure interface
// and then move the implementation to the manager types.
// Eg, Geth has its accounts, different from ErisMint

package types

import (
	"github.com/tendermint/ed25519"
	. "github.com/tendermint/go-common"
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
	return Fmt("PrivAccount{%X}", pA.Address)
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
		PanicSanity(Fmt("Expected 64 bytes but got %v", len(privKeyBytes)))
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
