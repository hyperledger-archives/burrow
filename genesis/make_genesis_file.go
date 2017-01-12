// Copyright 2015-2017 Monax Industries (UK) Ltd.
// This file is part of the Eris platform (Eris)

// Eris is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris is distributed WITHOUT ANY WARRANTY pursuant to
// the terms of the Gnu General Public Licence, version 3, including
// (but not limited to) Clause 15 thereof. See the text of the
// GNU General Public License, version 3 for full terms.

// You should have received a copy of the GNU General Public License
// along with Eris.  If not, see <http://www.gnu.org/licenses/>.

package genesis

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/eris-ltd/common/go/common"
	ptypes "github.com/eris-ltd/eris-db/permission/types"

	"github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
)

//------------------------------------------------------------------------------------
// core functions

func GenerateKnown(chainID, accountsPathCSV, validatorsPathCSV string) (string, error) {
	var genDoc *GenesisDoc

	// TODO [eb] eliminate reading priv_val ... [zr] where?
	if accountsPathCSV == "" || validatorsPathCSV == "" {
		return "", fmt.Errorf("both accounts.csv and validators.csv is required")
	}

	pubkeys, amts, names, perms, setbits, err := parseCsv(validatorsPathCSV)
	if err != nil {
		return "", err
	}

	pubkeysA, amtsA, namesA, permsA, setbitsA, err := parseCsv(accountsPathCSV)
	if err != nil {
		return "", err
	}

	genDoc = newGenDoc(chainID, len(pubkeys), len(pubkeysA))
	for i, pk := range pubkeys {
		genDocAddValidator(genDoc, pk, amts[i], names[i], perms[i], setbits[i], i)
	}
	for i, pk := range pubkeysA {
		genDocAddAccount(genDoc, pk, amtsA[i], namesA[i], permsA[i], setbitsA[i], i)
	}

	buf, buf2, n := new(bytes.Buffer), new(bytes.Buffer), new(int)
	wire.WriteJSON(genDoc, buf, n, &err)
	if err != nil {
		return "", err
	}
	if err := json.Indent(buf2, buf.Bytes(), "", "\t"); err != nil {
		return "", err
	}

	return buf2.String(), nil
}

//-----------------------------------------------------------------------------
// gendoc convenience functions

func newGenDoc(chainID string, nVal, nAcc int) *GenesisDoc {
	genDoc := GenesisDoc{
		ChainID: chainID,
		// GenesisTime: time.Now(),
	}
	genDoc.Accounts = make([]GenesisAccount, nAcc)
	genDoc.Validators = make([]GenesisValidator, nVal)
	return &genDoc
}

func genDocAddAccount(genDoc *GenesisDoc, pubKey crypto.PubKeyEd25519, amt int64, name string, perm, setbit ptypes.PermFlag, index int) {
	addr := pubKey.Address()
	acc := GenesisAccount{
		Address: addr,
		Amount:  amt,
		Name:    name,
		Permissions: &ptypes.AccountPermissions{
			Base: ptypes.BasePermissions{
				Perms:  perm,
				SetBit: setbit,
			},
		},
	}
	if index < 0 {
		genDoc.Accounts = append(genDoc.Accounts, acc)
	} else {
		genDoc.Accounts[index] = acc
	}
}

func genDocAddValidator(genDoc *GenesisDoc, pubKey crypto.PubKeyEd25519, amt int64, name string, perm, setbit ptypes.PermFlag, index int) {
	addr := pubKey.Address()
	genDoc.Validators[index] = GenesisValidator{
		PubKey: pubKey,
		Amount: amt,
		Name:   name,
		UnbondTo: []BasicAccount{
			{
				Address: addr,
				Amount:  amt,
			},
		},
	}
	// [zr] why no index < 0 like in genDocAddAccount?
}

//-----------------------------------------------------------------------------
// util functions

// convert hex strings to ed25519 pubkeys
func pubKeyStringsToPubKeys(pubkeys []string) ([]crypto.PubKeyEd25519, error) {
	pubKeys := make([]crypto.PubKeyEd25519, len(pubkeys))
	for i, k := range pubkeys {
		pubBytes, err := hex.DecodeString(k)
		if err != nil {
			return pubKeys, err
		}
		copy(pubKeys[i][:], pubBytes)
	}
	return pubKeys, nil
}

// empty is over written
func ifExistsElse(list []string, index int, defaultValue string) string {
	if len(list) > index {
		if list[index] != "" {
			return list[index]
		}
	}
	return defaultValue
}


