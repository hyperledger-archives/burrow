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

package genesis

import (
	"fmt"

	ptypes "github.com/hyperledger/burrow/permission/types"

	"github.com/tendermint/go-crypto"
)

const (
	PublicKeyEd25519ByteLength   int = 32
	PublicKeySecp256k1ByteLength int = 64
)

// NewGenesisAccount returns a new GenesisAccount
func NewGenesisAccount(address []byte, amount int64, name string,
	permissions *ptypes.AccountPermissions) *GenesisAccount {
	return &GenesisAccount{
		Address:     address,
		Amount:      amount,
		Name:        name,
		Permissions: permissions,
	}
}

func NewGenesisValidator(amount int64, name string, unbondToAddress []byte,
	unbondAmount int64, keyType string, publicKeyBytes []byte) (*GenesisValidator, error) {
	// convert the key bytes into a typed fixed size byte array
	var typedPublicKeyBytes []byte
	switch keyType {
	case "ed25519":
		// TODO: [ben] functionality and checks need to be inherit in the type
		if len(publicKeyBytes) != PublicKeyEd25519ByteLength {
			return nil, fmt.Errorf("Invalid length provided for ed25519 public key (%v bytes provided but expected %v bytes)",
				len(publicKeyBytes), PublicKeyEd25519ByteLength)
		}
		// ed25519 has type byte 0x01
		typedPublicKeyBytes = make([]byte, PublicKeyEd25519ByteLength+1)
		// prepend type byte to public key
		typedPublicKeyBytes = append([]byte{crypto.TypeEd25519}, publicKeyBytes...)
	case "secp256k1":
		if len(publicKeyBytes) != PublicKeySecp256k1ByteLength {
			return nil, fmt.Errorf("Invalid length provided for secp256k1 public key (%v bytes provided but expected %v bytes)",
				len(publicKeyBytes), PublicKeySecp256k1ByteLength)
		}
		// secp256k1 has type byte 0x02
		typedPublicKeyBytes = make([]byte, PublicKeySecp256k1ByteLength+1)
		// prepend type byte to public key
		typedPublicKeyBytes = append([]byte{crypto.TypeSecp256k1}, publicKeyBytes...)
	default:
		return nil, fmt.Errorf("Unsupported key type (%s)", keyType)
	}
	newPublicKey, err := crypto.PubKeyFromBytes(typedPublicKeyBytes)
	if err != nil {
		return nil, err
	}
	// ability to unbond to multiple accounts currently unused
	var unbondTo []BasicAccount

	return &GenesisValidator{
		PubKey: newPublicKey,
		Amount: unbondAmount,
		Name:   name,
		UnbondTo: append(unbondTo, BasicAccount{
			Address: unbondToAddress,
			Amount:  unbondAmount,
		}),
	}, nil
}
