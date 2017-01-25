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
	"encoding/csv"
	"fmt"
	"os"

	ptypes "github.com/eris-ltd/eris-db/permission/types"
)

// parseCsvIntoValidators is a helper function to read a csv file in the following format:
// >> pubkey, starting balance, name, permissions, setbit
// and returns the records as a slice of GenesisValidator
func parseCsvIntoValidators(filePath string) ([]ValidatorAccount, error) {

	params, err := readCsv(filePath)
}

// parseCsvIntoAccounts is a helper function to read a csv file in the following format:
// >> pubkey, starting balance, name, permissions, setbit
// and returns the records as a slice of GenesisAccount
// func parseCsvIntoAccounts(filePath string) (pubKeys []crypto.PubKeyEd25519, amts []int64, names []string, perms, setbits []ptypes.PermFlag, err error) {
func parseCsvIntoAccounts(filePath string) ([]GenesisAccount, error) {


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

// readGenesisAccountsRecord takes a slice of strings of the format:
// 
func readGenesisAccountsRecord([]string) (GenesisAccount, error) {


	address := publicKey.Address()
	genesisAccount := GenesisAccount{
		Address: address,
		Amount:  amount,
		Name:    name,
		Permissions: &ptypes.AccountPermissions{
			Base: ptypes.BasePermissions{
				Perms:  permissions,
				SetBit: setbit,
			}
		}
	}
	return genesisAccount, nil
}


// readCsv is a helper function to load the Csv file
func readCsv(filePath string) ([][]string, GenesisError) {
	csvFile, err := os.Open(filePath)
	if err != nil {
		return nil, GenesisError{fmt.Errorf("Couldn't open file %s: %v", filePath, err)}
	}
	defer csvFile.Close()

	r := csv.NewReader(csvFile)
	//r.FieldsPerRecord = # of records expected
	params, err := r.ReadAll()
	if err != nil {
		return nil, GenesisError{fmt.Errorf("Couldn't read file: %v", err)}
	}
	return params, nil
}
