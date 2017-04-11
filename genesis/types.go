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
	"os"
	"time"

	ptypes "github.com/hyperledger/burrow/permission/types"

	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

//------------------------------------------------------------
// we store the GenesisDoc in the db under this key

var GenDocKey = []byte("GenDocKey")

//------------------------------------------------------------
// core types for a genesis definition

type BasicAccount struct {
	Address []byte `json:"address"`
	Amount  int64  `json:"amount"`
}

type GenesisAccount struct {
	Address     []byte                     `json:"address"`
	Amount      int64                      `json:"amount"`
	Name        string                     `json:"name"`
	Permissions *ptypes.AccountPermissions `json:"permissions"`
}

type GenesisValidator struct {
	PubKey   crypto.PubKey  `json:"pub_key"`
	Amount   int64          `json:"amount"`
	Name     string         `json:"name"`
	UnbondTo []BasicAccount `json:"unbond_to"`
}

// GenesisPrivateValidator marshals the state of the private
// validator for the purpose of Genesis creation; and hence
// is defined in genesis and not under consensus, where
// PrivateValidator (currently inherited from Tendermint) is.
type GenesisPrivateValidator struct {
	Address    string        `json:"address"`
	PubKey     []interface{} `json:"pub_key"`
	PrivKey    []interface{} `json:"priv_key"`
	LastHeight int64         `json:"last_height"`
	LastRound  int64         `json:"last_round"`
	LastStep   int64         `json:"last_step"`
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

//------------------------------------------------------------
// Make genesis state from file

func GenesisDocFromJSON(jsonBlob []byte) (genState *GenesisDoc) {
	var err error
	wire.ReadJSONPtr(&genState, jsonBlob, &err)
	if err != nil {
		fmt.Printf("Couldn't read GenesisDoc: %v", err)
		// TODO: on error return error, not exit
		os.Exit(1)
	}
	return
}

//------------------------------------------------------------
// Methods for genesis types
// NOTE: breaks formatting convention
// TODO: split each genesis type in its own file definition

//------------------------------------------------------------
// GenesisAccount methods

// Clone clones the genesis account
func (genesisAccount *GenesisAccount) Clone() GenesisAccount {
	// clone the address
	addressClone := make([]byte, len(genesisAccount.Address))
	copy(addressClone, genesisAccount.Address)
	// clone the account permissions
	accountPermissionsClone := genesisAccount.Permissions.Clone()
	return GenesisAccount{
		Address:     addressClone,
		Amount:      genesisAccount.Amount,
		Name:        genesisAccount.Name,
		Permissions: &accountPermissionsClone,
	}
}

//------------------------------------------------------------
// GenesisValidator methods

// Clone clones the genesis validator
func (genesisValidator *GenesisValidator) Clone() (GenesisValidator, error) {
	if genesisValidator == nil {
		return GenesisValidator{}, fmt.Errorf("Cannot clone nil GenesisValidator.")
	}
	if genesisValidator.PubKey == nil {
		return GenesisValidator{}, fmt.Errorf("Invalid GenesisValidator %s with nil public key.",
			genesisValidator.Name)
	}
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
	}, nil
}

//------------------------------------------------------------
// BasicAccount methods

// Clone clones the basic account
func (basicAccount *BasicAccount) Clone() BasicAccount {
	// clone the address
	addressClone := make([]byte, len(basicAccount.Address))
	copy(addressClone, basicAccount.Address)
	return BasicAccount{
		Address: addressClone,
		Amount:  basicAccount.Amount,
	}
}
