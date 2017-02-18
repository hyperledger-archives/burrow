// Copyright 2015-2017 Monax Industries Limited.
// This file is part of the Monax platform (Monax)

// Monax is free software: you can use, redistribute it and/or modify
// it only under the terms of the GNU General Public License, version
// 3, as published by the Free Software Foundation.

// Monax is distributed WITHOUT ANY WARRANTY pursuant to
// the terms of the Gnu General Public Licence, version 3, including
// (but not limited to) Clause 15 thereof. See the text of the
// GNU General Public License, version 3 for full terms.

// You should have received a copy of the GNU General Public License,
// version 3, with Monax.  If not, see <http://www.gnu.org/licenses/>.

package genesis

import (
	"fmt"

	ptypes "github.com/eris-ltd/eris-db/permission/types"

	"github.com/tendermint/go-crypto"
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
	var publicKey crypto.PubKey
	// convert the key bytes into a typed fixed size byte array
	switch keyType {
	case "ed25519":
		// TODO: [ben] functionality and checks need to be inherit in the type
		if len(publicKeyBytes) != 32 {
			return nil, fmt.Errorf("Invalid length provided for ed25519 public key (len %v)",
				len(publicKeyBytes))
		}
		publicKey := crypto.PubKeyEd25519{}
		copy(publicKey[:], publicKeyBytes[:32])
	case "secp256k1":
		if len(publicKeyBytes) != 64 {
			return nil, fmt.Errorf("Invalid length provided for secp256k1 public key (len %v)",
				len(publicKeyBytes))
		}
		publicKey := crypto.PubKeySecp256k1{}
		copy(publicKey[:], publicKeyBytes[:64])
	default:
		return nil, fmt.Errorf("Unsupported key type (%s)", keyType)
	}

	// ability to unbond to multiple accounts currently unused
	var unbondTo []BasicAccount

	return &GenesisValidator{
		PubKey: publicKey,
		Amount: amount,
		Name:   name,
		UnbondTo: append(unbondTo, BasicAccount{
			Address: unbondToAddress,
			Amount:  unbondAmount,
		}),
	}, nil
}
