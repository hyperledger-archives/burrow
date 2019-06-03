// Copyright 2019 Monax Industries Limited
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

package crypto

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	tmEd25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tmSecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
)

func PublicKeyFromTendermintPubKey(pubKey tmCrypto.PubKey) (PublicKey, error) {
	switch pk := pubKey.(type) {
	case tmEd25519.PubKeyEd25519:
		return PublicKeyFromBytes(pk[:], CurveTypeEd25519)
	case tmSecp256k1.PubKeySecp256k1:
		return PublicKeyFromBytes(pk[:], CurveTypeSecp256k1)
	default:
		return PublicKey{}, fmt.Errorf("unrecognised tendermint public key type: %v", pk)
	}

}
func PublicKeyFromABCIPubKey(pubKey abci.PubKey) (PublicKey, error) {
	switch pubKey.Type {
	case CurveTypeEd25519.ABCIType():
		return PublicKey{
			CurveType: CurveTypeEd25519,
			PublicKey: pubKey.Data,
		}, nil
	case CurveTypeSecp256k1.ABCIType():
		return PublicKey{
			CurveType: CurveTypeEd25519,
			PublicKey: pubKey.Data,
		}, nil
	}
	return PublicKey{}, fmt.Errorf("did not recognise ABCI PubKey type: %s", pubKey.Type)
}

// PublicKey extensions

// Return the ABCI PubKey. See Tendermint protobuf.go for the go-crypto conversion this is based on
func (p PublicKey) ABCIPubKey() abci.PubKey {
	return abci.PubKey{
		Type: p.CurveType.ABCIType(),
		Data: p.PublicKey,
	}
}

func (p PublicKey) TendermintPubKey() tmCrypto.PubKey {
	switch p.CurveType {
	case CurveTypeEd25519:
		pk := tmEd25519.PubKeyEd25519{}
		copy(pk[:], p.PublicKey)
		return pk
	case CurveTypeSecp256k1:
		pk := tmSecp256k1.PubKeySecp256k1{}
		copy(pk[:], p.PublicKey)
		return pk
	default:
		return nil
	}
}

// Signature extensions

func (sig Signature) TendermintSignature() []byte {
	return sig.Signature
}
