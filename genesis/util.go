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
)

// parseCsvIntoValidators is a helper function to read a csv file in the following format:
// >> pubkey, starting balance, name, permissions, setbit
// and returns the records as a slice of GenesisValidator

// parseCsvIntoAccounts is a helper function to read a csv file in the following format:
// >> pubkey, starting balance, name, permissions, setbit
// and returns the records as a slice of GenesisAccount
// func parseCsvIntoAccounts(filePath string) (pubKeys []crypto.PubKeyEd25519, amts []int64, names []string, perms, setbits []ptypes.PermFlag, err error) {
func parseCsvIntoAccounts(filePath string) ([]GenesisAccount, error) {

	csvFile, err := os.Open(filePath)
	if err != nil {
		common.Exit(fmt.Errorf("Couldn't open file: %s: %v", filePath, err))
	}
	defer csvFile.Close()

	r := csv.NewReader(csvFile)
	//r.FieldsPerRecord = # of records expected
	params, err := r.ReadAll()
	if err != nil {
		common.Exit(fmt.Errorf("Couldn't read file: %v", err))

	}

	pubkeys := make([]string, len(params))
	amtS := make([]string, len(params))
	names = make([]string, len(params))
	permsS := make([]string, len(params))
	setbitS := make([]string, len(params))
	for i, each := range params {
		pubkeys[i] = each[0]
		amtS[i] = ifExistsElse(each, 1, "1000")
		names[i] = ifExistsElse(each, 2, "")
		permsS[i] = ifExistsElse(each, 3, fmt.Sprintf("%d", ptypes.DefaultPermFlags))
		setbitS[i] = ifExistsElse(each, 4, permsS[i])
	}

	//TODO convert int to uint64, see issue #25
	perms = make([]ptypes.PermFlag, len(permsS))
	for i, perm := range permsS {
		pflag, err := strconv.Atoi(perm)
		if err != nil {
			common.Exit(fmt.Errorf("Permissions must be an integer"))
		}
		perms[i] = ptypes.PermFlag(pflag)
	}
	setbits = make([]ptypes.PermFlag, len(setbitS))
	for i, setbit := range setbitS {
		setbitsFlag, err := strconv.Atoi(setbit)
		if err != nil {
			common.Exit(fmt.Errorf("SetBits must be an integer"))
		}
		setbits[i] = ptypes.PermFlag(setbitsFlag)
	}

	// convert amts to ints
	amts = make([]int64, len(amtS))
	for i, a := range amtS {
		if amts[i], err = strconv.ParseInt(a, 10, 64); err != nil {
			err = fmt.Errorf("Invalid amount: %v", err)
			return
		}
	}

	// convert pubkey hex strings to struct
	pubKeys, err = pubKeyStringsToPubKeys(pubkeys)
	if err != nil {
		return
	}

	return pubKeys, amts, names, perms, setbits, nil
}