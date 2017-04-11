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
	"time"

	ptypes "github.com/hyperledger/burrow/permission/types"
	wire "github.com/tendermint/go-wire"
)

// MakeGenesisDocFromAccounts takes a chainName and a slice of pointers to GenesisAccount,
// and a slice of pointers to GenesisValidator to construct a GenesisDoc, or returns an error on
// failure.  In particular MakeGenesisDocFromAccount uses the local time as a
// timestamp for the GenesisDoc.
func MakeGenesisDocFromAccounts(chainName string, accounts []*GenesisAccount,
	validators []*GenesisValidator) (GenesisDoc, error) {

	// TODO: assert valid accounts and validators
	// TODO: [ben] expose setting global permissions
	globalPermissions := ptypes.DefaultAccountPermissions.Clone()
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
		genesisValidatorCopy, err := genesisValidator.Clone()
		if err != nil {
			return GenesisDoc{}, err
		}
		validatorsCopy[i] = genesisValidatorCopy
	}
	genesisDoc := GenesisDoc{
		GenesisTime: time.Now(),
		// TODO: this needs to be corrected for ChainName, and ChainId
		// is the derived hash from the GenesisDoc serialised bytes
		ChainID:    chainName,
		Params:     genesisParameters,
		Accounts:   accountsCopy,
		Validators: validatorsCopy,
	}
	return genesisDoc, nil
}

// GetGenesisFileBytes returns the JSON (not-yet) canonical bytes for a given
// GenesisDoc or an error.  In a first rewrite, rely on go-wire
// for the JSON serialisation with type-bytes.
func GetGenesisFileBytes(genesisDoc *GenesisDoc) ([]byte, error) {

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
