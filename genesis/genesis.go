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
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"sort"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

// How many bytes to take from the front of the GenesisDoc hash to append
// to the ChainName to form the ChainID. The idea is to avoid some classes
// of replay attack between chains with the same name.
const ChainIDHashSuffixBytes = 4

// we store the GenesisDoc in the db under this key

var GenDocKey = []byte("GenDocKey")

//------------------------------------------------------------
// core types for a genesis definition

type BasicAccount struct {
	Address acm.Address `json:"address"`
	Amount  uint64      `json:"amount"`
}

type GenesisAccount struct {
	BasicAccount
	Name        string                    `json:"name"`
	Permissions ptypes.AccountPermissions `json:"permissions"`
}

type GenesisValidator struct {
	// Address  is convenient to have in file for reference, but otherwise ignored since derived from PubKey
	Address  acm.Address    `json:"address,omitempty"`
	Amount   uint64         `json:"amount"`
	PubKey   crypto.PubKey  `json:"pub_key"`
	Name     string         `json:"name"`
	UnbondTo []BasicAccount `json:"unbond_to"`
}

type GenesisParams struct {
	GlobalPermissions ptypes.AccountPermissions `json:"global_permissions"`
}

//------------------------------------------------------------
// GenesisDoc is stored in the state database

type GenesisDoc struct {
	GenesisTime time.Time          `json:"genesis_time"`
	ChainName   string             `json:"chain_name"`
	Salt        []byte             `json:"salt,omitempty"`
	Params      GenesisParams      `json:"params"`
	Accounts    []GenesisAccount   `json:"accounts"`
	Validators  []GenesisValidator `json:"validators"`
}

// JSONBytes returns the JSON (not-yet) canonical bytes for a given
// GenesisDoc or an error.  In a first rewrite, rely on go-wire
// for the JSON serialisation with type-bytes.
func (genesisDoc *GenesisDoc) JSONBytes() ([]byte, error) {
	// TODO: write JSON in canonical order
	var err error
	buffer, n := new(bytes.Buffer), new(int)
	// write JSON with go-wire type-bytes (for public keys)
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

func (genesisDoc *GenesisDoc) Hash() []byte {
	genesisDocBytes, err := genesisDoc.JSONBytes()
	if err != nil {
		panic(fmt.Errorf("could not create hash of GenesisDoc: %v", err))
	}
	hasher := sha256.New()
	hasher.Write(genesisDocBytes)
	return hasher.Sum(nil)
}

func (genesisDoc *GenesisDoc) ChainID() string {
	return fmt.Sprintf("%s-%X", genesisDoc.ChainName, genesisDoc.Hash()[:ChainIDHashSuffixBytes])
}

//------------------------------------------------------------
// Make genesis state from file

func GenesisDocFromJSON(jsonBlob []byte) (*GenesisDoc, error) {
	var err error
	genDoc := new(GenesisDoc)
	wire.ReadJSONPtr(genDoc, jsonBlob, &err)
	if err != nil {
		return nil, fmt.Errorf("couldn't read GenesisDoc: %v", err)
	}
	return genDoc, nil
}

//------------------------------------------------------------
// Methods for genesis types
// NOTE: breaks formatting convention
// TODO: split each genesis type in its own file definition

//------------------------------------------------------------
// GenesisAccount methods

// Clone clones the genesis account
func (genesisAccount *GenesisAccount) Clone() GenesisAccount {
	// clone the account permissions
	return GenesisAccount{
		BasicAccount: BasicAccount{
			Address: genesisAccount.Address,
			Amount:  genesisAccount.Amount,
		},
		Name:        genesisAccount.Name,
		Permissions: genesisAccount.Permissions.Clone(),
	}
}

//------------------------------------------------------------
// GenesisValidator methods

func (gv *GenesisValidator) Validator() acm.Validator {
	return acm.ConcreteValidator{
		Address: acm.MustAddressFromBytes(gv.PubKey.Address()),
		PubKey:  gv.PubKey,
		Power:   uint64(gv.Amount),
	}.Validator()
}

// Clone clones the genesis validator
func (gv *GenesisValidator) Clone() GenesisValidator {
	// clone the addresses to unbond to
	unbondToClone := make([]BasicAccount, len(gv.UnbondTo))
	for i, basicAccount := range gv.UnbondTo {
		unbondToClone[i] = basicAccount.Clone()
	}
	return GenesisValidator{
		PubKey:   gv.PubKey,
		Amount:   gv.Amount,
		Name:     gv.Name,
		UnbondTo: unbondToClone,
	}
}

//------------------------------------------------------------
// BasicAccount methods

// Clone clones the basic account
func (basicAccount *BasicAccount) Clone() BasicAccount {
	return BasicAccount{
		Address: basicAccount.Address,
		Amount:  basicAccount.Amount,
	}
}

// MakeGenesisDocFromAccounts takes a chainName and a slice of pointers to GenesisAccount,
// and a slice of pointers to GenesisValidator to construct a GenesisDoc, or returns an error on
// failure.  In particular MakeGenesisDocFromAccount uses the local time as a
// timestamp for the GenesisDoc.
func MakeGenesisDocFromAccounts(chainName string, salt []byte, genesisTime time.Time, accounts map[string]acm.Account,
	validators map[string]acm.Validator) *GenesisDoc {

	// Establish deterministic order of accounts by name so we obtain identical GenesisDoc
	// from identical input
	names := make([]string, 0, len(accounts))
	for name := range accounts {
		names = append(names, name)
	}
	sort.Strings(names)
	// copy slice of pointers to accounts into slice of accounts
	genesisAccounts := make([]GenesisAccount, 0, len(accounts))
	for _, name := range names {
		acc := accounts[name]
		genesisAccounts = append(genesisAccounts, GenesisAccount{
			Name:        name,
			Permissions: acc.Permissions(),
			BasicAccount: BasicAccount{
				Address: acc.Address(),
				Amount:  acc.Balance(),
			},
		})
	}
	// Sigh...
	names = names[:0]
	for name := range validators {
		names = append(names, name)
	}
	sort.Strings(names)
	// copy slice of pointers to validators into slice of validators
	genesisValidators := make([]GenesisValidator, 0, len(validators))
	for _, name := range names {
		val := validators[name]
		genesisValidators = append(genesisValidators, GenesisValidator{
			Name:    name,
			Address: val.Address(),
			PubKey:  val.PubKey(),
			Amount:  val.Power(),
			// Simpler to just do this by convention
			UnbondTo: []BasicAccount{
				{
					Amount:  val.Power(),
					Address: val.Address(),
				},
			},
		})
	}
	return &GenesisDoc{
		ChainName:   chainName,
		Salt:        salt,
		GenesisTime: genesisTime,
		Params: GenesisParams{
			GlobalPermissions: permission.DefaultAccountPermissions.Clone(),
		},
		Accounts:   genesisAccounts,
		Validators: genesisValidators,
	}
}
