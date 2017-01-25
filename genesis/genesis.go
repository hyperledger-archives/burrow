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
	"bytes"
	"encoding/json"
	"time"

	ptypes "github.com/eris-ltd/eris-db/permission/types"
	wire "github.com/tendermint/go-wire"
)

// MakeGenesisDocFromAccounts takes a chainName and a slice of GenesisAccount,
// and a slice of GenesisValidator to construct a GenesisDoc, or returns an error on
// failure.  In particular MakeGenesisDocFromAccount uses the local time as a
// timestamp for the GenesisDoc.
func MakeGenesisDocFromAccounts(chainName string, accounts []GenesisAccount, validators []GenesisValidator) (GenesisDoc, error) {

	// TODO: assert valid accounts and validators
	genesisParameters := GenesisParams{
		GlobalPermissions: ptypes.DefaultAccountPermissions,
	}
	genesisDoc := GenesisDoc{
		GenesisTime: time.Now(),
		// TODO: this needs to be corrected for ChainName, and ChainId
		// is the derived hash from the GenesisDoc serialised bytes
		ChainID:    chainName,
		Params:     genesisParameters,
		Accounts:   accounts,
		Validators: validators,
	}
	return genesisDoc, nil
}

// GetGenesisFileBytes returns the JSON (not-yet) canonical bytes for a given
// GenesisDoc or an error.  In a first rewrite, rely on go-wire
// for the JSON serialisation with type-bytes.
func GetGenesisFileBytes(genesisDoc *GenesisDoc) ([]byte, error) {

	// TODO: write JSON in canonical order

	buffer, n, err := new(bytes.Buffer), new(int), new(error)
	// write JSON with go-wire type-bytes (for public keys); deprecate
	wire.WriteJSON(genesisDoc, buffer, n, &err)
	if err != nil {
		return nil, err
	}
	// rewrite buffer with indentation
	indentedBuffer := new(bytes.Buffer)
	if err := json.Indent(indentedBuffer, buffer.Bytes(), "", "\t"); err != nil {
		return nil, err
	}
	return indentedBuffer.Bytes(), nil
}
