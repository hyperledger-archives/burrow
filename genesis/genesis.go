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

)

// MakeGenesisDocFromAccounts takes a chainName and a slice of GenesisAccount,
// and a slice of GenesisValidator to construct a GenesisDoc, or returns an error on
// failure.  In particular MakeGenesisDocFromAccount uses the local time as a
// timestamp for the GenesisDoc.
func MakeGenesisDocFromAccounts(chainName string, accounts []GenesisAccount, validators []GenesisValidator) (GenesisDoc, error) {
	genesisDoc := GenesisDoc {
		// TODO: this needs to be corrected for ChainName, and ChainId
		// is the derived hash from the GenesisDoc
		ChainID: chainName,
		GenesisTime: time.Now(),
	}
}

// GetGenesisFileBytes returns the JSON canonical bytes for a given
// GenesisDoc or an error.
func GetGenesisFileBytes(genesisDoc *GenesisDoc) ([]bytes, error) {

}

func 