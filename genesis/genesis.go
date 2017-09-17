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
	"encoding/json"
	"fmt"
	"time"

	"crypto/sha256"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
)

// we store the GenesisDoc in the db under this key

var GenDocKey = []byte("GenDocKey")

//------------------------------------------------------------
// core types for a genesis definition

type BasicAccount struct {
	Address account.Address `json:"address"`
	Amount  int64           `json:"amount"`
}

type GenesisAccount struct {
	BasicAccount
	Name        string                     `json:"name"`
	Permissions *ptypes.AccountPermissions `json:"permissions"`
}

type GenesisValidator struct {
	PubKey   crypto.PubKey  `json:"pub_key"`
	Amount   int64          `json:"amount"`
	Name     string         `json:"name"`
	UnbondTo []BasicAccount `json:"unbond_to"`
}

type GenesisParams struct {
	GlobalPermissions *ptypes.AccountPermissions `json:"global_permissions"`
}

//------------------------------------------------------------
// GenesisDoc is stored in the state database

type GenesisDoc struct {
	GenesisTime time.Time          `json:"genesis_time"`
	ChainID     string             `json:"chain_id"`
	Params      *GenesisParams     `json:"params"`
	Accounts    []GenesisAccount   `json:"accounts"`
	Validators  []GenesisValidator `json:"validators"`
}

// GenesisFileBytes returns the JSON (not-yet) canonical bytes for a given
// GenesisDoc or an error.  In a first rewrite, rely on go-wire
// for the JSON serialisation with type-bytes.
func (genesisDoc *GenesisDoc) GenesisFileBytes() ([]byte, error) {
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
	genesisDocBytes, err := genesisDoc.GenesisFileBytes()
	if err != nil {
		panic(fmt.Errorf("could not create hash of GenesisDoc: %v", err))
	}
	hasher := sha256.New()
	hasher.Write(genesisDocBytes)
	return hasher.Sum(nil)
}

//------------------------------------------------------------
// Make genesis state from file

func GenesisDocFromJSON(jsonBlob []byte) (*GenesisDoc, error) {
	var err error
	genState := new(GenesisDoc)
	wire.ReadJSONPtr(genState, jsonBlob, &err)
	if err != nil {
		return nil, fmt.Errorf("couldn't read GenesisDoc: %v", err)
	}
	return genState, nil
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
	accountPermissionsClone := genesisAccount.Permissions.Clone()
	return GenesisAccount{
		BasicAccount: BasicAccount{
			Address: genesisAccount.Address,
			Amount:  genesisAccount.Amount,
		},
		Name:        genesisAccount.Name,
		Permissions: &accountPermissionsClone,
	}
}

//------------------------------------------------------------
// GenesisValidator methods

// Clone clones the genesis validator
func (genesisValidator *GenesisValidator) Clone() GenesisValidator {
	// clone the addresses to unbond to
	unbondToClone := make([]BasicAccount, len(genesisValidator.UnbondTo))
	for i, basicAccount := range genesisValidator.UnbondTo {
		unbondToClone[i] = basicAccount.Clone()
	}
	return GenesisValidator{
		PubKey:   genesisValidator.PubKey,
		Amount:   genesisValidator.Amount,
		Name:     genesisValidator.Name,
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
func MakeGenesisDocFromAccounts(chainID string, accounts []*GenesisAccount,
	validators []*GenesisValidator) *GenesisDoc {

	globalPermissions := permission.DefaultAccountPermissions.Clone()
	genesisParameters := &GenesisParams{
		GlobalPermissions: &globalPermissions,
	}
	// copy slice of pointers to accounts into slice of accounts
	accountsCopy := make([]GenesisAccount, len(accounts))
	for i, genesisAccount := range accounts {
		accountsCopy[i] = genesisAccount.Clone()
	}
	// copy slice of pointers to validators into slice of validators
	validatorsCopy := make([]GenesisValidator, len(validators))
	for i, genesisValidator := range validators {
		genesisValidatorCopy := genesisValidator.Clone()
		validatorsCopy[i] = genesisValidatorCopy
	}
	return &GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     chainID,
		Params:      genesisParameters,
		Accounts:    accountsCopy,
		Validators:  validatorsCopy,
	}
}
